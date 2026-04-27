package rpc

import (
	"github.com/DemonZack/simplejrpc-go"
	"github.com/DemonZack/simplejrpc-go/core"
	"github.com/DemonZack/simplejrpc-go/core/config"
	"github.com/DemonZack/simplejrpc-go/core/gi18n"
	"github.com/DemonZack/simplejrpc-go/net/gsock"
	"github.com/DemonZack/simplejrpc-go/os/gpath"
	"os"
)

type Server struct {
}

// NewServer 创建RPC服务器实例
func NewServer() *Server {
	return &Server{}
}

// RegisterHandles 注册所有RPC处理函数
func (s *Server) RegisterHandles(ds interface {
	RegisterHandle(name string, handler func(*gsock.Request) (any, error), middlewares ...gsock.RPCMiddleware)
},
) {
	// 注册基础接口
	ds.RegisterHandle("ping", s.Ping)

	// Connection Management
	ds.RegisterHandle("db.testConnection", s.TestConnection)
	ds.RegisterHandle("db.connect", s.Connect)
	ds.RegisterHandle("db.disconnect", s.Disconnect)
	ds.RegisterHandle("db.listConnections", s.ListConnections)

	// Connection Configuration (Saved Connections)
	ds.RegisterHandle("db.saveConnectionConfig", s.SaveConnectionConfig)
	ds.RegisterHandle("db.listConnectionConfigs", s.ListConnectionConfigs)
	ds.RegisterHandle("db.getConnectionConfig", s.GetConnectionConfig)
	ds.RegisterHandle("db.deleteConnectionConfig", s.DeleteConnectionConfig)

	// Database Operations
	ds.RegisterHandle("db.listDatabases", s.ListDatabases)
	ds.RegisterHandle("db.getDatabaseInfo", s.GetDatabaseInfo)
	ds.RegisterHandle("db.createDatabase", s.CreateDatabase)
	ds.RegisterHandle("db.dropDatabase", s.DropDatabase)
	ds.RegisterHandle("db.switchDatabase", s.SwitchDatabase)
	ds.RegisterHandle("db.exportDatabase", s.ExportDatabase)
	ds.RegisterHandle("db.importDatabase", s.ImportDatabase)
	ds.RegisterHandle("db.listDatabaseDumpTasks", s.ListDatabaseDumpTasks)
	ds.RegisterHandle("db.getDatabaseDumpTask", s.GetDatabaseDumpTask)
	ds.RegisterHandle("db.getDatabaseDumpTaskLogs", s.GetDatabaseDumpTaskLogs)
	ds.RegisterHandle("db.cancelDatabaseDumpTask", s.CancelDatabaseDumpTask)

	// Schema Operations (PostgreSQL only)
	ds.RegisterHandle("db.listSchemas", s.ListSchemas)

	// Table Operations
	ds.RegisterHandle("db.listTables", s.ListTables)
	ds.RegisterHandle("db.getTableSchema", s.GetTableSchema)
	ds.RegisterHandle("db.getTableData", s.GetTableData)

	// Table Designer (DDL)
	ds.RegisterHandle("db.createTable", s.CreateTable)
	ds.RegisterHandle("db.alterTable", s.AlterTable)
	ds.RegisterHandle("db.dropTable", s.DropTable)
	ds.RegisterHandle("db.renameTable", s.RenameTable)
	ds.RegisterHandle("db.getTableDDL", s.GetTableDDL)

	// Query Operations
	ds.RegisterHandle("db.executeSQL", s.ExecuteSQL)

	// Data Modification
	ds.RegisterHandle("db.updateRow", s.UpdateRow)
	ds.RegisterHandle("db.insertRow", s.InsertRow)
	ds.RegisterHandle("db.deleteRow", s.DeleteRow)
	ds.RegisterHandle("db.batchModify", s.BatchModify)

	// Import / Export
	ds.RegisterHandle("db.exportTable", s.ExportTable)
	ds.RegisterHandle("db.exportTableDDL", s.ExportTableDDL)
	ds.RegisterHandle("db.importCSV", s.ImportCSV)
	ds.RegisterHandle("db.importTableSQL", s.ImportTableSQL)

}

// Start 启动RPC服务器
func (s *Server) Start(sockPath string) error {
	_ = os.MkdirAll(sockPath, 0o755)
	ds := simplejrpc.NewDefaultServer(
		gsock.WithJsonRpcSimpleServiceHandler(gsock.NewJsonRpcSimpleServiceHandler()),
	)
	gpath.GmCfgPath = "./"
	gi18n.Instance().SetPath("./i18n")
	core.InitContainer(config.WithConfigEnvFormatterOptionFunc("test"))

	s.RegisterHandles(ds)

	return ds.StartServer(sockPath)
}

// Ping 心跳检测
func (s *Server) Ping(req *gsock.Request) (any, error) {
	// core.Container.Log().Info("收到Ping请求")
	return "pong", nil
}
