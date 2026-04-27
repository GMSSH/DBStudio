const C = {
  brand6: '#5772FF',
  brand5: '#708BFF',
  brand7: '#3F53D9',
  white100: '#FFFFFF',
  white85: 'rgba(255,255,255,0.85)',
  white65: 'rgba(255,255,255,0.65)',
  white45: 'rgba(255,255,255,0.45)',
  white35: 'rgba(255,255,255,0.35)',
  white25: 'rgba(255,255,255,0.25)',
  white20: 'rgba(255,255,255,0.20)',
  white15: 'rgba(255,255,255,0.15)',
  white12: 'rgba(255,255,255,0.12)',
  white10: 'rgba(255,255,255,0.10)',
  white8: 'rgba(255,255,255,0.08)',
  white4: 'rgba(255,255,255,0.04)',
  white0: 'rgba(255,255,255,0)',
  formBg: 'rgba(255,255,255,0.06)',
  overlayBg: 'rgba(22,28,40,0.74)',
  overlayBgStrong: 'rgba(18,24,36,0.88)',
  overlayBorder: 'rgba(255,255,255,0.16)',
  overlayShadow: '0 28px 72px rgba(2,7,16,0.44), 0 12px 30px rgba(2,7,16,0.28), inset 0 1px 0 rgba(255,255,255,0.12)',
  red6: '#DA323F'
}

const S = {
  fontFamily: "'Source Han Sans SC','Noto Sans SC','Microsoft YaHei',system-ui,sans-serif",
  fontSize: '14px',
  fontSizeSm: '13px',
  fontSizeLg: '16px',
  fontWeightMedium: '500',
  fontWeightStrong: '600',
  radiusSm: '4px',
  radiusMd: '6px',
  radiusLg: '8px'
}

export const themeOverrides = {
  common: {
    primaryColor: C.brand6,
    primaryColorHover: C.brand5,
    primaryColorPressed: C.brand7,
    primaryColorSuppl: C.brand6,
    fontFamily: S.fontFamily,
    fontSize: S.fontSize,
    fontWeightStrong: S.fontWeightStrong,
    textColor1: C.white100,
    textColor2: C.white85,
    textColor3: C.white45,
    placeholderColor: C.white35,
    borderColor: C.white15,
    borderRadius: S.radiusMd,
    borderRadiusSmall: S.radiusSm,
    heightSmall: '26px',
    heightMedium: '34px',
    heightLarge: '40px',
    bodyColor: 'transparent'
  },
  Button: {
    borderRadiusMedium: S.radiusMd,
    borderRadiusSmall: S.radiusSm,
    borderRadiusLarge: S.radiusLg,
    colorPrimary: C.brand6,
    colorHoverPrimary: C.brand5,
    colorPressedPrimary: C.brand7,
    colorFocusPrimary: C.brand6,
    colorDisabledPrimary: C.brand6,
    textColorPrimary: C.white100,
    textColorHoverPrimary: C.white100,
    textColorPressedPrimary: C.white100,
    textColorFocusPrimary: C.white100,
    textColorDisabledPrimary: C.white25,
    borderPrimary: '1px solid transparent',
    borderHoverPrimary: '1px solid transparent',
    borderPressedPrimary: '1px solid transparent',
    borderFocusPrimary: '1px solid transparent',
    borderDisabledPrimary: '1px solid transparent',
    colorDefault: C.white0,
    colorHoverDefault: C.white8,
    colorPressedDefault: C.white4,
    colorFocusDefault: C.white0,
    colorDisabledDefault: C.white0,
    textColorDefault: C.white65,
    textColorHoverDefault: C.white85,
    textColorPressedDefault: C.white65,
    textColorFocusDefault: C.white65,
    textColorDisabledDefault: C.white25,
    borderDefault: `1px solid ${C.white25}`,
    borderHoverDefault: `1px solid ${C.brand6}`,
    borderPressedDefault: `1px solid ${C.brand5}`,
    borderFocusDefault: `1px solid ${C.brand6}`,
    borderDisabledDefault: `1px solid ${C.white12}`,
    colorGhost: C.white0,
    colorHoverGhost: C.white8,
    colorPressedGhost: C.white4,
    colorFocusGhost: C.white0,
    colorDisabledGhost: C.white0,
    textColorGhost: C.white65,
    textColorHoverGhost: C.white85,
    textColorPressedGhost: C.white65,
    textColorFocusGhost: C.white65,
    textColorDisabledGhost: C.white25,
    borderGhost: `1px solid ${C.white15}`,
    borderHoverGhost: `1px solid ${C.white25}`,
    borderPressedGhost: `1px solid ${C.white10}`,
    borderFocusGhost: `1px solid ${C.brand6}`,
    borderDisabledGhost: `1px solid ${C.white8}`
  },
  Input: {
    borderRadius: S.radiusMd,
    color: C.formBg,
    colorFocus: C.formBg,
    colorDisabled: 'rgba(255,255,255,0.03)',
    border: `1px solid ${C.white15}`,
    borderHover: `1px solid ${C.white45}`,
    borderFocus: `1px solid ${C.brand6}`,
    borderDisabled: `1px solid ${C.white12}`,
    boxShadowFocus: '0 0 0 3px rgba(87,114,255,0.28)',
    textColor: C.white85,
    textColorDisabled: C.white25,
    placeholderColor: C.white35,
    iconColor: C.white65,
    iconColorHover: C.white100,
    iconColorDisabled: C.white25,
    clearColor: C.white65,
    clearColorHover: C.white100,
    clearColorPressed: C.white65
  },
  Select: {
    peers: {
      InternalSelection: {
        borderRadius: S.radiusMd,
        color: C.formBg,
        colorActive: C.formBg,
        border: `1px solid ${C.white15}`,
        borderActive: `1px solid ${C.brand6}`,
        borderFocus: `1px solid ${C.brand6}`,
        borderHover: `1px solid ${C.white45}`,
        boxShadowActive: '0 0 0 3px rgba(87,114,255,0.28)',
        textColor: C.white85,
        placeholderColor: C.white35,
        arrowColor: C.white65
      },
      InternalSelectMenu: {
        color: C.overlayBgStrong,
        borderRadius: S.radiusLg,
        boxShadow: '0 18px 38px rgba(2,7,16,0.36), inset 0 1px 0 rgba(255,255,255,0.08)',
        border: `1px solid ${C.overlayBorder}`,
        optionTextColor: C.white85,
        optionTextColorActive: C.white100,
        optionColorPending: C.white8,
        optionColorActive: 'rgba(87,114,255,0.18)'
      }
    }
  },
  Form: {
    labelFontWeight: S.fontWeightMedium,
    labelFontSizeMedium: S.fontSizeSm,
    labelTextColor: C.white65,
    asteriskColor: C.red6
  },
  DataTable: {
    borderRadius: S.radiusLg,
    thColor: 'rgba(255,255,255,0.025)',
    tdColor: 'transparent',
    borderColor: C.white10,
    thColorHover: C.white8,
    tdColorHover: 'rgba(87,114,255,0.07)',
    tdColorStriped: 'rgba(255,255,255,0.016)',
    thTextColor: C.white45,
    tdTextColor: C.white85
  },
  Layout: {
    color: 'transparent',
    siderColor: 'transparent',
    headerColor: 'transparent'
  },
  Card: {
    color: C.white4,
    borderColor: C.white10,
    borderRadius: '10px',
    boxShadow: '0 4px 20px rgba(0,0,0,0.3)',
    titleTextColor: C.white85,
    textColor: C.white65
  },
  Modal: {
    color: C.overlayBg,
    boxShadow: C.overlayShadow
  },
  Dialog: {
    color: C.overlayBg,
    titleTextColor: C.white100,
    textColor: C.white65,
    borderRadius: '18px',
    boxShadow: C.overlayShadow
  },
  Drawer: {
    color: C.overlayBg,
    borderRadius: '20px',
    boxShadow: C.overlayShadow
  },
  Tabs: {
    tabColor: 'transparent',
    tabColorSegment: C.white8,
    tabBorderColor: C.white10,
    tabTextColorActiveLine: C.brand6,
    tabTextColorHoverLine: C.white85,
    barColor: C.brand6
  },
  Tree: {
    nodeColor: 'transparent',
    nodeColorHover: 'rgba(87,114,255,0.10)',
    nodeColorPressed: 'rgba(87,114,255,0.15)',
    nodeTextColor: C.white65,
    nodeTextColorActive: C.white100
  },
  Dropdown: {
    color: 'rgba(24,24,32,0.96)',
    borderRadius: '10px',
    boxShadow: '0 8px 32px rgba(0,0,0,0.5), 0 2px 8px rgba(0,0,0,0.24), inset 0 1px 0 rgba(255,255,255,0.06)',
    optionTextColor: C.white85,
    optionColorHover: C.white8,
    optionColorActive: 'rgba(87,114,255,0.16)',
    optionTextColorActive: C.brand5,
    dividerColor: C.white8
  },
  Progress: {
    railColor: 'rgba(255,255,255,0.08)',
    railColorError: 'rgba(218,50,63,0.14)',
    railColorSuccess: 'rgba(77,204,115,0.14)',
    fillColor: '#5268EF',
    fillColorError: '#DA323F',
    fillColorSuccess: '#4DCC73',
    textColor: C.white65,
    textColorCircle: C.white65
  }
}
