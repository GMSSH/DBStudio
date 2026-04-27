package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// dataDir 是运行时动态计算的数据目录，通过 getDataDir() 获取。
// 布局：二进制在 .../app/bin/<exe>，往上5层到 /.__gmssh/，再进 tmp/sqlmanager
var (
	dataDirOnce sync.Once
	dataDirPath string
)

func getDataDir() string {
	dataDirOnce.Do(func() {
		exe, err := os.Executable()
		if err != nil {
			// fallback：使用当前工作目录
			cwd, _ := os.Getwd()
			dataDirPath = filepath.Join(cwd, "data")
		} else {
			// /.__gmssh/plugin/xiaojun/dbmanager/app/bin/<exe>
			// 上5级 → /.__gmssh/ → tmp/sqlmanager
			binDir := filepath.Dir(exe)
			dataDirPath = filepath.Clean(filepath.Join(binDir, "../../../../../tmp/sqlmanager"))
		}
		if err := os.MkdirAll(dataDirPath, 0755); err != nil {
			log.Printf("[ConnectionStore] WARNING: cannot create data directory %s: %v", dataDirPath, err)
		}
		log.Printf("[ConnectionStore] data directory: %s", dataDirPath)
	})
	return dataDirPath
}

// DataDir returns the runtime data directory used by DBManager.
func DataDir() string {
	return getDataDir()
}

func encryptionFilePath() string {
	return filepath.Join(getDataDir(), ".dbmanager_connections.enc")
}
func encryptionFileBakPath() string {
	return filepath.Join(getDataDir(), ".dbmanager_connections.enc.bak")
}
func keyFilePath() string {
	return filepath.Join(getDataDir(), ".dbmanager_secret")
}

// SavedConnection represents a saved database connection configuration
type SavedConnection struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	DBType    string    `json:"dbType"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // Will be encrypted
	Database  string    `json:"database"`
	FilePath  string    `json:"filePath"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ConnectionStore manages encrypted connection configurations
type ConnectionStore struct {
	mu            sync.RWMutex
	encryptionKey []byte
}

var (
	globalStore *ConnectionStore
	storeOnce   sync.Once
)

// GetConnectionStore returns the global connection store instance
func GetConnectionStore() *ConnectionStore {
	storeOnce.Do(func() {
		key, err := loadOrCreateEncryptionKey()
		if err != nil {
			// 理论上不应发生（随机生成不会失败），panic 让问题显现
			panic(fmt.Sprintf("[ConnectionStore] 无法加载或创建加密密钥: %v", err))
		}
		globalStore = &ConnectionStore{
			encryptionKey: key,
		}
	})
	return globalStore
}

// loadOrCreateEncryptionKey 按优先级获取加密密钥：
//  1. 环境变量 DB_ENCRYPTION_KEY（64位十六进制，适合 K8s Secret / 多实例）
//  2. 本地密钥文件 .dbmanager_secret（首次启动随机生成并持久化）
func loadOrCreateEncryptionKey() ([]byte, error) {
	// 优先级 1：环境变量
	if envKey := os.Getenv("DB_ENCRYPTION_KEY"); envKey != "" {
		key, err := hex.DecodeString(envKey)
		if err == nil && len(key) == 32 {
			log.Println("[ConnectionStore] using environment variable DB_ENCRYPTION_KEY")
			return key, nil
		}
		log.Println("[ConnectionStore] WARNING: DB_ENCRYPTION_KEY is invalid (expect 64-char hex), ignoring")
	}

	// 优先级 2：读取本地密钥文件
	kf := keyFilePath()
	if data, err := os.ReadFile(kf); err == nil {
		key, err := hex.DecodeString(string(data))
		if err == nil && len(key) == 32 {
			return key, nil
		}
		// 文件内容损坏，重新生成
		log.Printf("[ConnectionStore] WARNING: key file %s is invalid, regenerating", kf)
	}

	// 首次启动：随机生成32字节密钥并持久化
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("生成随机密钥失败: %w", err)
	}
	keyHex := hex.EncodeToString(key)
	if err := os.WriteFile(kf, []byte(keyHex), 0600); err != nil {
		return nil, fmt.Errorf("保存密钥文件失败: %w", err)
	}
	log.Printf("[ConnectionStore] first start, generated random encryption key at %s", kf)
	return key, nil
}

// encrypt encrypts data using AES-256-GCM
func (cs *ConnectionStore) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(cs.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decrypt decrypts data using AES-256-GCM
func (cs *ConnectionStore) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(cs.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// loadConnections loads and decrypts connections from file.
// If the file is corrupted or was encrypted with a different key, it backs up
// the old file and returns an empty map so new connections can be saved normally.
func (cs *ConnectionStore) loadConnections() (map[string]*SavedConnection, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	enc := encryptionFilePath()
	bak := encryptionFileBakPath()

	// Check if file exists
	if _, err := os.Stat(enc); os.IsNotExist(err) {
		return make(map[string]*SavedConnection), nil
	}

	// Read encrypted file
	encryptedData, err := os.ReadFile(enc)
	if err != nil {
		return nil, fmt.Errorf("failed to read connections file: %w", err)
	}

	if len(encryptedData) == 0 {
		return make(map[string]*SavedConnection), nil
	}

	// Decrypt data
	decryptedData, err := cs.decrypt(encryptedData)
	if err != nil {
		log.Printf("[ConnectionStore] WARNING: cannot decrypt connections file (%v). "+
			"Backing up to %s and starting fresh.\n", err, bak)
		_ = os.Rename(enc, bak)
		return make(map[string]*SavedConnection), nil
	}

	// Parse JSON
	var connections map[string]*SavedConnection
	if err := json.Unmarshal(decryptedData, &connections); err != nil {
		log.Printf("[ConnectionStore] WARNING: cannot parse connections file (%v). "+
			"Backing up to %s and starting fresh.\n", err, bak)
		_ = os.Rename(enc, bak)
		return make(map[string]*SavedConnection), nil
	}

	return connections, nil
}

// saveConnections encrypts and saves connections to file
func (cs *ConnectionStore) saveConnections(connections map[string]*SavedConnection) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Serialize to JSON
	data, err := json.Marshal(connections)
	if err != nil {
		return fmt.Errorf("failed to serialize connections: %w", err)
	}

	// Encrypt data
	encryptedData, err := cs.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt connections: %w", err)
	}

	// Write to file with restrictive permissions
	if err := os.WriteFile(encryptionFilePath(), encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write connections file: %w", err)
	}

	return nil
}

// SaveConnection saves or updates a connection configuration
func (cs *ConnectionStore) SaveConnection(conn *SavedConnection) error {
	connections, err := cs.loadConnections()
	if err != nil {
		return err
	}

	// Generate ID if new
	if conn.ID == "" {
		conn.ID = generateID()
		conn.CreatedAt = time.Now()
	}
	conn.UpdatedAt = time.Now()

	connections[conn.ID] = conn

	return cs.saveConnections(connections)
}

// ListConnections returns all saved connections (without passwords)
func (cs *ConnectionStore) ListConnections() ([]*SavedConnection, error) {
	connections, err := cs.loadConnections()
	if err != nil {
		return nil, err
	}

	result := make([]*SavedConnection, 0, len(connections))
	for _, conn := range connections {
		// Create copy without password
		safeCopy := *conn
		safeCopy.Password = "" // Don't expose password in list
		result = append(result, &safeCopy)
	}

	return result, nil
}

// GetConnection retrieves a single connection by ID (with password)
func (cs *ConnectionStore) GetConnection(id string) (*SavedConnection, error) {
	connections, err := cs.loadConnections()
	if err != nil {
		return nil, err
	}

	conn, exists := connections[id]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", id)
	}

	return conn, nil
}

// DeleteConnection removes a connection by ID
func (cs *ConnectionStore) DeleteConnection(id string) error {
	connections, err := cs.loadConnections()
	if err != nil {
		return err
	}

	if _, exists := connections[id]; !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	delete(connections, id)

	return cs.saveConnections(connections)
}

// generateID generates a unique ID for a connection
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
