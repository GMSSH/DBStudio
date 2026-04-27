interface MenuItems {
    label: string;
    icon: string;
    path: string;
    tooltip?: boolean;
    iconSize?: string;
    loading?: () => boolean | undefined;
    disabled?: () => boolean | undefined;
}
type __VLS_Props = {
    routerStore?: string;
    defaultPath?: string;
    menus: MenuItems[];
};
declare function __VLS_template(): {
    attrs: Partial<{}>;
    slots: {
        footer?(_: {}): any;
    };
    refs: {};
    rootEl: HTMLDivElement;
};
type __VLS_TemplateResult = ReturnType<typeof __VLS_template>;
declare const __VLS_component: import('vue').DefineComponent<__VLS_Props, {}, {}, {}, {}, import('vue').ComponentOptionsMixin, import('vue').ComponentOptionsMixin, {}, string, import('vue').PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import('vue').ComponentProvideOptions, false, {}, HTMLDivElement>;
declare const _default: __VLS_WithTemplateSlots<typeof __VLS_component, __VLS_TemplateResult["slots"]>;
export default _default;
type __VLS_WithTemplateSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
