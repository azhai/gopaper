(function () {
  function emptySiteMeta() {
    return {
      SITE_TITLE: '', SITE_DESC: '', SITE_THEME: '',
      NAV_ORDER: 0, NAV_HIDE: false, DIR_TYPE: '', LAYOUT: '',
      SORT_ORDER: '', ARTICLES_PER_PAGE: 10,
      HERO_TITLE: '', HERO_SUBTITLE: '', HERO_IMAGE: '',
      HERO_CTA_TEXT: '', HERO_CTA_LINK: '',
      FEATURE_TITLE: '', FEATURES: [],
      NEWS_TITLE: '', NEWS_LINK: '',
      CONTACT_EMAIL: '', CONTACT_PHONE: '', CONTACT_ADDRESS: '',
      FOOTER_TEXT: '', ICP: '',
    };
  }

  function cloneSiteMeta(meta) {
    return Object.assign({}, meta, {
      FEATURES: (meta.FEATURES || []).map(function (card) { return Object.assign({}, card); }),
    });
  }

  function emptyTheme() {
    return {
      id: '', name: '', enabledAt: '',
      layout: { container: 'default', radius: 'soft', header: 'gradient', hero: 'gradient', position: 'sticky', shadow: 'soft' },
      colors: { primary: '#d4a017', accent: '#e8833a', background: '#fffdf5', surface: '#ffffff', text: '#5d4e37', muted: '#8b7d6b', border: '#f0e6c0' },
      fonts: { body: 'sans', heading: 'sans', size: 16 },
    };
  }

  function cloneTheme(theme) {
    return Object.assign({}, theme, {
      layout: Object.assign({}, theme.layout),
      colors: Object.assign({}, theme.colors),
      fonts: Object.assign({}, theme.fonts),
    });
  }

  function emptyThemeOverview() {
    return { currentId: '', nextId: '', nextAt: '', defaultCss: '', themes: [], presets: [] };
  }

  function splitCategory(path) {
    var parts = (path || '').split('/').map(function (part) { return part.trim(); }).filter(Boolean);
    return [parts[0] || '', parts[1] || ''];
  }

  function contentHref(item) {
    return item.dirPath ? '/' + item.dirPath + '/' + item.slug : '/' + item.slug;
  }

  function sectionLabel(dirs, dirPath) {
    if (!dirPath) return '首页';
    var found = (dirs || []).find(function (d) { return d.dirPath === dirPath; });
    return found ? (found.title || dirPath) : dirPath;
  }

  function typeLabel(type) {
    if (type === 'page') return '单页';
    if (type === 'product') return '产品目录';
    if (type === 'news') return '文章列表';
    if (type === 'docs') return '文档列表';
    return type || '栏目';
  }

  function formatDate(date) { return date || '未设置'; }

  function textLength(text) { return (text || '').replace(/\s/g, '').length; }

  function today() { return new Date().toISOString().slice(0, 10); }

  function apiErrorMessage(res) {
    if (res && res.data && typeof res.data === 'object') {
      var first = Object.values(res.data).find(function (v) { return typeof v === 'string'; });
      if (typeof first === 'string') return first;
    }
    return (res && res.message) || '请检查填写内容';
  }

  function formatSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1024 / 1024).toFixed(1) + ' MB';
  }

  window.adminUI = {
    emptySiteMeta: emptySiteMeta,
    cloneSiteMeta: cloneSiteMeta,
    emptyTheme: emptyTheme,
    cloneTheme: cloneTheme,
    emptyThemeOverview: emptyThemeOverview,
    splitCategory: splitCategory,
    contentHref: contentHref,
    sectionLabel: sectionLabel,
    typeLabel: typeLabel,
    formatDate: formatDate,
    textLength: textLength,
    today: today,
    apiErrorMessage: apiErrorMessage,
    formatSize: formatSize,
  };
})();
