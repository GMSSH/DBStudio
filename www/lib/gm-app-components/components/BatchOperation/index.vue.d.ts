interface Props {
    /** 是否跳过确认弹窗，默认为 false */
    skipConfirm?: boolean;
}
/** 批量操作结果统计 */
interface BatchOperationResult {
    /** 成功数量 */
    success: number;
    /** 失败数量 */
    failed: number;
    /** 重复数量 */
    duplicate: number;
}
/** 批量操作配置 */
interface BatchOperationConfig {
    /** 操作标题 */
    title: string;
    /** 确认提示信息 */
    confirmMessage: string;
    /** 待处理的数据项列表 */
    items: any[];
    /** 单项操作处理函数 */
    operation: (item: any, index: number) => Promise<any>;
    /** 操作完成后的回调函数 */
    onComplete?: (results: BatchOperationResult) => void;
}
declare const _default: import('vue').DefineComponent<Props, {
    startBatchOperation: (config: BatchOperationConfig) => void;
}, {}, {}, {}, import('vue').ComponentOptionsMixin, import('vue').ComponentOptionsMixin, {
    resultDialogClosed: () => any;
}, string, import('vue').PublicProps, Readonly<Props> & Readonly<{
    onResultDialogClosed?: (() => any) | undefined;
}>, {
    skipConfirm: boolean;
}, {}, {}, {}, string, import('vue').ComponentProvideOptions, false, {}, any>;
export default _default;
