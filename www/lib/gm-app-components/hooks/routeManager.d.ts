type RouteParams = Record<string, any>;
interface RouteEntry {
    fullPath: string;
    route: string;
    params: RouteParams;
}
type NextRouteEntry = Omit<RouteEntry, "route" | "params"> & {
    route?: string;
    params?: RouteParams;
};
type NextFn = (action?: boolean | string | NextRouteEntry) => void;
type RouteGuard = (to: RouteEntry, from: RouteEntry | undefined, next: NextFn) => void | Promise<void>;
export declare const useRouteStore: (config?: string | {
    defaultRoute?: string;
    routeKey: string;
    otherConfig?: {
        [key: string]: any;
    };
}) => {
    readonly route: string | null;
    readonly fullPath: string | null;
    readonly currentFullRoute: string | null;
    readonly fromRoute: {
        fullPath: string;
        route: string;
        params: RouteParams;
    } | undefined;
    readonly toRoute: {
        fullPath: string;
        route: string;
        params: RouteParams;
    } | undefined;
    readonly currentIndex: number;
    getCurrentParams: () => RouteParams;
    setCurrentParams: (params: RouteParams) => void;
    setRoute: (path: string, params?: RouteParams, replaceCurrent?: boolean, force?: boolean) => void;
    goBack: (extraParams?: RouteParams) => Promise<void>;
    goNext: (extraParams?: RouteParams) => Promise<void>;
    canGoBack: () => boolean;
    canGoNext: () => boolean;
    goToIndex: (index: number, extraParams?: RouteParams) => Promise<void>;
    reset: () => void;
    printDebug: () => {
        fullPath: string;
        route: string;
        params: RouteParams;
    }[];
    getRoute: (options?: {
        level?: number;
    }) => {
        route: string;
        fullPath: string;
        params: RouteParams;
    } | null;
    beforeEach: (guard: RouteGuard) => void;
    afterEach: (guard: RouteGuard) => void;
};
export {};
