export interface ProgressTask {
    /** 唯一标识 */
    id: string;
    /** 弹窗标题 (顶部标题栏文字) */
    title: string;
    /** 应用名称 (图标下方文字) */
    appName: string;
    /** 应用图标 URL */
    appIcon: string;
    /** 当前状态 */
    status: 'pending' | 'running' | 'success' | 'error';
    /** 状态文字（成功/失败时显示的描述） */
    statusText: string;
    /** 进度百分比 (0-100)，由假进度条自动推进 */
    progress: number;
    /** 进度条下方左侧的文字 */
    message: string;
    /** 日志内容（纯文本字符串，由调用方通过轮询接口以覆盖方式写入） */
    logData: string;
    /** 是否展示日志面板 */
    showLogs: boolean;
    /** 日志按钮文字（如 "安装日志"，默认 "日志"） */
    logBtnText: string;
    /** 弹窗是否可见（关闭只隐藏弹窗，不影响外部进程） */
    visible: boolean;
    /** 预计耗时（秒），用于控制假进度条的推进速率 */
    estimatedDuration: number;
    /** 创建时间戳 */
    createdAt: number;
    /** 内部：真实的精确进度（带小数），用于均匀累加 */
    _exactProgress: number;
    /** 内部：假进度定时器 ID */
    _progressTimer: number | null;
}
/**
 * 添加一个进度任务并自动启动假进度条，返回任务 id。
 * 弹窗默认打开 (visible: true)。
 */
declare function addTask(config?: {
    title?: string;
    appName?: string;
    appIcon?: string;
    message?: string;
    logBtnText?: string;
    /** 预计耗时（秒），用于让进度条更接近真实时间。不传则使用默认速率。 */
    estimatedDuration?: number;
}): string;
/**
 * 更新指定任务的属性（日志以覆盖方式更新，每次传入完整 logData）
 */
declare function updateTask(id: string, data: Partial<Pick<ProgressTask, 'title' | 'appName' | 'appIcon' | 'status' | 'statusText' | 'progress' | 'message' | 'logData' | 'showLogs' | 'logBtnText'>>): void;
/**
 * 完成任务（成功或失败），进度跳到 100%，停止假进度条。
 */
declare function finishTask(id: string, result: {
    success: boolean;
    statusText?: string;
}): void;
/**
 * 显示弹窗（仅控制 UI 可见性，不影响任务状态和外部进程）
 */
declare function showTask(id: string): void;
/**
 * 隐藏弹窗（仅关闭 UI，不停止外部进程，也不移除任务）
 */
declare function hideTask(id: string): void;
/**
 * 切换日志面板显示/隐藏
 */
declare function toggleLogs(id: string): void;
/**
 * 彻底移除任务（同时清理定时器），用于流程完全结束后的清理。
 */
declare function removeTask(id: string): void;
/**
 * 清除所有任务
 */
declare function clearAll(): void;
/**
 * 获取进度管理器（composable 风格）
 */
export declare function useProgress(): {
    /** 响应式任务列表（Ref） */
    tasks: globalThis.Ref<{
        id: string;
        title: string;
        appName: string;
        appIcon: string;
        status: "pending" | "running" | "success" | "error";
        statusText: string;
        progress: number;
        message: string;
        logData: string;
        showLogs: boolean;
        logBtnText: string;
        visible: boolean;
        estimatedDuration: number;
        createdAt: number;
        _exactProgress: number;
        _progressTimer: number | null;
    }[], ProgressTask[] | {
        id: string;
        title: string;
        appName: string;
        appIcon: string;
        status: "pending" | "running" | "success" | "error";
        statusText: string;
        progress: number;
        message: string;
        logData: string;
        showLogs: boolean;
        logBtnText: string;
        visible: boolean;
        estimatedDuration: number;
        createdAt: number;
        _exactProgress: number;
        _progressTimer: number | null;
    }[]>;
    addTask: typeof addTask;
    updateTask: typeof updateTask;
    finishTask: typeof finishTask;
    showTask: typeof showTask;
    hideTask: typeof hideTask;
    toggleLogs: typeof toggleLogs;
    removeTask: typeof removeTask;
    clearAll: typeof clearAll;
};
export {};
