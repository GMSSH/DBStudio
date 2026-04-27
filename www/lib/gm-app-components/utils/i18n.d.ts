export class I18nParse {
    constructor({ dirPath, localesPath, addPath }: {
        dirPath: any;
        localesPath: any;
        addPath: any;
    });
    chineseRegex: RegExp;
    init(): Promise<void>;
    readDirectory(directoryPath: any): Promise<void>;
    readFileByType(filePath: any): Promise<void>;
    logger(filePath: any, result: any): Promise<void>;
    getCurrentTime(): {
        dirName: string;
        fileName: string;
    };
    parseJsOrTs(content: any): any;
    parseVue(content: any): string;
    replaceChineseInJS(scriptText: any): any;
    recursiveHandle(str: any, expressions: any): any;
    extractTemplateExpressions(str: any): any[];
    replaceChineseInTags(content: any): any;
    tagContentPosition(left: any, right: any, token: any): any;
    replaceChineseInAttrs(content: any): any;
    mergeJson(target: any, source: any): any;
    createRandomKey(value: any): string | undefined;
    findKeyWithValue(obj: any, value: any, path?: string): any;
    findValueWithKey(obj: any, value: any): any;
    generateUniqueString(length: any): string;
    removeComments(content: any): any;
    restoreComments(content: any): any;
    #private;
}
