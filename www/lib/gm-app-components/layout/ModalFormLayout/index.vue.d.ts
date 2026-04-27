declare function __VLS_template(): {
    attrs: Partial<{}>;
    slots: {
        default?(_: {}): any;
    };
    refs: {};
    rootEl: HTMLDivElement;
};
type __VLS_TemplateResult = ReturnType<typeof __VLS_template>;
declare const __VLS_component: import('vue').DefineComponent<globalThis.ExtractPropTypes<{
    title: {
        type: StringConstructor;
        default: string;
    };
    width: {
        type: NumberConstructor;
        default: number;
    };
    confirmText: {
        type: StringConstructor;
        default: string;
    };
    cancelText: {
        type: StringConstructor;
        default: string;
    };
    showBtn: {
        type: BooleanConstructor;
        default: boolean;
    };
    paddingStyle: {
        type: StringConstructor;
        default: string;
    };
    globLoading: {
        type: BooleanConstructor;
        default: boolean;
    };
    okLoading: {
        type: BooleanConstructor;
        default: boolean;
    };
}>, {}, {}, {}, {}, import('vue').ComponentOptionsMixin, import('vue').ComponentOptionsMixin, {
    onConfirm: (...args: any[]) => void;
    onCancel: (...args: any[]) => void;
}, string, import('vue').PublicProps, Readonly<globalThis.ExtractPropTypes<{
    title: {
        type: StringConstructor;
        default: string;
    };
    width: {
        type: NumberConstructor;
        default: number;
    };
    confirmText: {
        type: StringConstructor;
        default: string;
    };
    cancelText: {
        type: StringConstructor;
        default: string;
    };
    showBtn: {
        type: BooleanConstructor;
        default: boolean;
    };
    paddingStyle: {
        type: StringConstructor;
        default: string;
    };
    globLoading: {
        type: BooleanConstructor;
        default: boolean;
    };
    okLoading: {
        type: BooleanConstructor;
        default: boolean;
    };
}>> & Readonly<{
    onOnConfirm?: ((...args: any[]) => any) | undefined;
    onOnCancel?: ((...args: any[]) => any) | undefined;
}>, {
    title: string;
    width: number;
    confirmText: string;
    cancelText: string;
    showBtn: boolean;
    paddingStyle: string;
    globLoading: boolean;
    okLoading: boolean;
}, {}, {}, {}, string, import('vue').ComponentProvideOptions, true, {}, HTMLDivElement>;
declare const _default: __VLS_WithTemplateSlots<typeof __VLS_component, __VLS_TemplateResult["slots"]>;
export default _default;
type __VLS_WithTemplateSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
