type __VLS_Props = {
    showEditTop?: boolean;
    str: string;
    height: string;
    mode: string;
    showPlaceholder?: boolean;
};
declare function __VLS_template(): {
    attrs: Partial<{}>;
    slots: {
        placeholder?(_: {}): any;
    };
    refs: {
        editor: HTMLDivElement;
    };
    rootEl: HTMLDivElement;
};
type __VLS_TemplateResult = ReturnType<typeof __VLS_template>;
declare const __VLS_component: import('vue').DefineComponent<__VLS_Props, {}, {}, {}, {}, import('vue').ComponentOptionsMixin, import('vue').ComponentOptionsMixin, {
    "update:str": (...args: any[]) => void;
}, string, import('vue').PublicProps, Readonly<__VLS_Props> & Readonly<{
    "onUpdate:str"?: ((...args: any[]) => any) | undefined;
}>, {
    showEditTop: boolean;
    showPlaceholder: boolean;
}, {}, {}, {}, string, import('vue').ComponentProvideOptions, false, {
    editor: HTMLDivElement;
}, HTMLDivElement>;
declare const _default: __VLS_WithTemplateSlots<typeof __VLS_component, __VLS_TemplateResult["slots"]>;
export default _default;
type __VLS_WithTemplateSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
