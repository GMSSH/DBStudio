export type RequestFunction<T, P> = (params?: P) => Promise<T>;
type CallbackFunction<T> = (response: T) => void;
type ErrorCallbackFunction = (error: any) => boolean | Promise<boolean>;
type StopCallbackFunction = (reason: "error" | "manual" | "max_retries" | "component_unmounted", error?: any) => void;
export declare function usePoller<T, P = void>(): {
    poll: (config: {
        req: RequestFunction<T, P>;
        delay: number;
        callback?: CallbackFunction<T>;
        params?: P;
        options?: {
            immediate?: boolean;
            maxRetries?: number;
            onError?: ErrorCallbackFunction;
            onStop?: StopCallbackFunction;
        };
    }) => void;
    stop: () => void;
    isPolling: globalThis.Ref<boolean, boolean>;
};
/** 🔥 清空所有轮询 */
export declare function stopAllPollers(): void;
export {};
