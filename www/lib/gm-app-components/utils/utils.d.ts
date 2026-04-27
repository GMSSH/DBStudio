export declare const locale: string;
/**
 * 随机字符串
 * 包含大小写字母、数字和特殊符号的字符串
 */
export declare function generateRandomString(): string;
/**
 * 复制到剪切板
 */
export declare const copyToClipboard: (text: string) => boolean;
/**
 * 唤起文件选择窗口，并获取选择的文件
 * @returns
 */
export declare function getWinFiles(): Promise<unknown>;
/**
 *获取根元素css变量
 * @export
 * @param {string} cssVar css根元素变量
 * @return {string}
 */
export declare function useRootElementCssVariable(cssVar: string): string;
