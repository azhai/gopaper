document.addEventListener('alpine:init', function () {
  var CONTAINERS = [
    { value: 'narrow', label: '紧凑（880px）' },
    { value: 'default', label: '标准（1120px）' },
    { value: 'wide', label: '宽松（1320px）' },
    { value: 'full', label: '超宽（1560px）' },
  ];
  var RADII = [
    { value: 'sharp', label: '直角' },
    { value: 'soft', label: '柔和' },
    { value: 'round', label: '大圆角' },
    { value: 'pill', label: '圆润' },
  ];
  var BANDS = [
    { value: 'gradient', label: '渐变' },
    { value: 'solid', label: '纯色' },
    { value: 'plain', label: '素净' },
  ];
  var POSITIONS = [
    { value: 'sticky', label: '滚动时固定' },
    { value: 'static', label: '随页面滚动' },
  ];
  var SHADOWS = [
    { value: 'flat', label: '扁平（无阴影）' },
    { value: 'soft', label: '轻柔' },
    { value: 'strong', label: '明显' },
  ];
  var FONTS = [
    { value: 'sans', label: '无衬线（系统黑体）' },
    { value: 'serif', label: '衬线（宋体 / Georgia）' },
    { value: 'kai', label: '楷体' },
    { value: 'mono', label: '等宽' },
  ];
  var SIZES = [14, 15, 16, 17, 18, 19, 20];
  var COLOR_FIELDS = [
    { key: 'primary', label: '主色', hint: '链接、按钮、导航强调' },
    { key: 'accent', label: '强调色', hint: '横幅渐变与悬停' },
    { key: 'background', label: '页面底色', hint: '整站背景' },
    { key: 'surface', label: '卡片底色', hint: '内容区块背景' },
    { key: 'text', label: '正文色', hint: '标题与正文' },
    { key: 'muted', label: '次要文字', hint: '摘要、日期' },
    { key: 'border', label: '边框色', hint: '卡片与分隔线' },
  ];
  var HEX_RE = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i;

  function labelOf(options, value) {
    var found = options.find(function (option) { return option.value === value; });
    return found ? found.label.replace(/（.*/, '') : value;
  }

  function normalizeHex(value) {
    var hex = (value || '').trim().toLowerCase();
    if (!HEX_RE.test(hex)) return '';
    if (hex.length === 4) return '#' + hex[1] + hex[1] + hex[2] + hex[2] + hex[3] + hex[3];
    return hex;
  }

  function nowDateTime() {
    var now = new Date();
    function pad(n) { return String(n).padStart(2, '0'); }
    return now.getFullYear() + '-' + pad(now.getMonth() + 1) + '-' + pad(now.getDate()) + 'T' + pad(now.getHours()) + ':' + pad(now.getMinutes());
  }

  function previewDocument(css) {
    return '<!DOCTYPE html><html lang="zh-CN"><head><meta charset="utf-8">' +
      '<meta name="viewport" content="width=device-width,initial-scale=1">' +
      '<link rel="stylesheet" href="/static/css/style.css"><style>' + css + '</style>' +
      '</head><body>' +
      '<header class="site-header"><div class="container header-inner">' +
      '<span class="logo">你的品牌</span>' +
      '<nav class="main-nav"><a href="#" class="active">首页</a><a href="#">产品</a><a href="#">关于我们</a></nav>' +
      '</div></header>' +
      '<section class="hero"><div class="container hero-inner">' +
      '<h1>让好产品被更多人看见</h1>' +
      '<p class="hero-subtitle">一句话说明你能为客户做什么</p>' +
      '<a href="#" class="btn btn-primary">联系我们</a>' +
      '</div></section>' +
      '<section class="features"><div class="container">' +
      '<h2 class="section-title">产品与服务</h2>' +
      '<div class="features-grid">' +
      '<div class="feature-card"><h3>企业建站</h3><p>三天上线，内容自己在后台改，不需要懂代码。</p><a href="#" class="card-link">查看详情</a></div>' +
      '<div class="feature-card"><h3>技术咨询</h3><p>架构评审、性能优化与安全加固，按次或按月合作。</p><a href="#" class="card-link">查看详情</a></div>' +
      '<div class="feature-card"><h3>定制开发</h3><p>从表单到小程序，把重复的手工活变成自动化。</p><a href="#" class="card-link">查看详情</a></div>' +
      '</div></div></section>' +
      '<section class="list-section"><div class="container"><div class="list-grid">' +
      '<div class="list-card"><div class="list-meta">2026-03-18</div><h3><a href="#">新版产品目录上线</a></h3></div>' +
      '<div class="list-card"><div class="list-meta">2026-03-02</div><h3><a href="#">我们开始提供技术顾问服务</a></h3></div>' +
      '</div></div></section>' +
      '<footer class="site-footer"><div class="container"><div class="footer-bottom-inner">' +
      '<p>© 2026 你的公司 All rights reserved.</p><p class="icp">京ICP备XXXXXXXX号</p>' +
      '</div></div></footer>' +
      '</body></html>';
  }

  Alpine.data('themesPage', function () {
    return {
      icon: adminIcon,
      loading: true,
      error: null,
      overview: adminUI.emptyThemeOverview(),
      draft: adminUI.emptyTheme(),
      editing: false,
      editingId: '',
      dirty: false,
      saving: false,
      busyId: '',
      previewCss: '',
      previewLabel: '当前网站',
      previewingDraft: false,
      previewTimer: 0,
      containers: CONTAINERS,
      radii: RADII,
      bands: BANDS,
      positions: POSITIONS,
      shadows: SHADOWS,
      fonts: FONTS,
      sizes: SIZES,
      colorFields: COLOR_FIELDS,
      init: function () {
        this.bindLifecycle();
        this.load();
      },
      bindLifecycle: function () {
        var self = this;
        this.$store.app.setGuard(function () { return self.dirty; }, '主题还有未保存的修改，离开后会丢失。');
        this.beforeUnload = function (e) {
          if (!self.dirty) return;
          e.preventDefault();
          e.returnValue = '';
        };
        window.addEventListener('beforeunload', this.beforeUnload);
      },
      destroy: function () {
        this.$store.app.clearGuard();
        window.removeEventListener('beforeunload', this.beforeUnload);
        window.clearTimeout(this.previewTimer);
      },
      get currentTheme() {
        var overview = this.overview;
        return overview.themes.find(function (theme) { return theme.id === overview.currentId; });
      },
      get upcomingTheme() {
        var overview = this.overview;
        return overview.themes.find(function (theme) { return theme.id === overview.nextId; });
      },
      get isNew() { return !this.editingId; },
      get previewDoc() { return previewDocument(this.previewCss); },
      get saveLabel() {
        if (this.saving) return '保存中…';
        return this.editingId ? '保存修改' : '保存主题';
      },
      readableTime: function (value) { return value ? value.replace('T', ' ') : '未安排'; },
      layoutSummary: function (theme) {
        return labelOf(CONTAINERS, theme.layout.container) + ' · ' + labelOf(RADII, theme.layout.radius) + ' · ' + labelOf(FONTS, theme.fonts.body);
      },
      fontLabel: function (value) {
        return value === 'sans' ? '无衬线' : '其他标题字体';
      },
      bandStyle: function (theme, part) {
        var style = part === 'header' ? theme.layout.header : theme.layout.hero;
        if (style === 'solid') return 'background:' + theme.colors.primary;
        if (style === 'plain') return 'background:' + theme.colors.surface + ';border-bottom:1px solid ' + theme.colors.border;
        return 'background:linear-gradient(90deg,' + theme.colors.primary + ',' + theme.colors.accent + ')';
      },
      swatchBody: function (theme) { return 'background:' + theme.colors.background; },
      swatchLine: function (theme, muted) { return 'background:' + (muted ? theme.colors.muted : theme.colors.text); },
      swatchPill: function (theme) { return 'background:' + theme.colors.primary; },
      hexValue: function (value) { return normalizeHex(value) || '#000000'; },
      load: async function () {
        this.loading = true;
        this.error = null;
        try {
          var res = await adminApi.getThemes();
          if (res.code !== 0 || !res.data) throw new Error(res.message || '无法读取主题');
          this.overview = res.data;
          if (!this.previewingDraft) this.showSavedPreview(res.data.currentId);
        } catch (err) {
          this.error = err instanceof Error ? err.message : '网络错误';
        } finally {
          this.loading = false;
        }
      },
      setDraft: function (mutate) {
        mutate(this.draft);
        this.dirty = true;
        this.previewingDraft = true;
        this.previewLabel = '正在编辑';
        this.scheduleDraftPreview();
      },
      updateLayout: function (key, value) {
        this.setDraft(function (theme) { theme.layout[key] = value; });
      },
      updateColor: function (key, value) {
        this.setDraft(function (theme) { theme.colors[key] = value; });
      },
      updateFont: function (key, value) {
        this.setDraft(function (theme) { theme.fonts[key] = value; });
      },
      updateFontSize: function (value) {
        this.setDraft(function (theme) { theme.fonts.size = parseInt(value, 10); });
      },
      updateName: function (value) {
        this.setDraft(function (theme) { theme.name = value; });
      },
      updateEnabledAt: function (value) {
        this.setDraft(function (theme) { theme.enabledAt = value.length > 16 ? value.slice(0, 16) : value; });
      },
      enableNow: function () {
        this.setDraft(function (theme) { theme.enabledAt = nowDateTime(); });
      },
      toggleSticky: function (checked) {
        this.setDraft(function (theme) { theme.layout.position = checked ? 'sticky' : 'static'; });
      },
      scheduleDraftPreview: function () {
        var self = this;
        window.clearTimeout(this.previewTimer);
        this.previewTimer = window.setTimeout(function () {
          adminApi.previewTheme(self.draft).then(function (res) {
            if (res.code === 0 && res.data) self.previewCss = res.data.css;
          });
        }, 220);
      },
      showSavedPreview: function (id) {
        window.clearTimeout(this.previewTimer);
        var theme = this.overview.themes.find(function (item) { return item.id === id; });
        this.previewingDraft = false;
        if (theme) {
          this.previewCss = theme.css || '';
          this.previewLabel = theme.name;
          return;
        }
        this.previewCss = this.overview.defaultCss;
        this.previewLabel = '内置默认样式';
      },
      startNew: function () {
        var self = this;
        this.guardDraft(function () {
          self.openEditor(adminUI.emptyTheme(), '', '正在编辑', '');
        });
      },
      editTheme: function (theme) {
        var self = this;
        this.guardDraft(function () {
          self.openEditor(adminUI.cloneTheme(theme), theme.id, '编辑 ' + theme.name, theme.css || '');
        });
      },
      usePreset: function (preset) {
        var self = this;
        var draft = adminUI.cloneTheme(preset);
        draft.enabledAt = '';
        this.guardDraft(function () {
          self.openEditor(draft, '', '基于「' + preset.name + '」', '');
        });
      },
      duplicateTheme: function (theme) {
        var self = this;
        var draft = adminUI.cloneTheme(theme);
        draft.name = (theme.name + ' 副本').slice(0, 40);
        draft.enabledAt = '';
        this.guardDraft(function () {
          self.openEditor(draft, '', '复制自「' + theme.name + '」', '');
        });
      },
      openEditor: function (draft, editingId, label, css) {
        this.draft = draft;
        this.editing = true;
        this.editingId = editingId;
        this.dirty = false;
        this.previewingDraft = true;
        this.previewLabel = label;
        if (css) {
          this.previewCss = css;
          return;
        }
        this.fetchPreview();
      },
      fetchPreview: function () {
        var self = this;
        adminApi.previewTheme(this.draft).then(function (res) {
          if (res.code === 0 && res.data) self.previewCss = res.data.css;
        });
      },
      guardDraft: function (open) {
        var self = this;
        if (!this.dirty) {
          open();
          return;
        }
        this.$store.overlay.confirm('放弃当前编辑？', '未保存的主题修改会丢失。', function () {
          self.dirty = false;
          open();
        });
      },
      closeEditor: function () {
        var self = this;
        this.guardDraft(function () {
          self.editing = false;
          self.editingId = '';
          self.draft = adminUI.emptyTheme();
          self.dirty = false;
          self.showSavedPreview(self.overview.currentId);
        });
      },
      saveDraft: async function () {
        if (this.saving) return;
        var name = this.draft.name.trim();
        if (!name) {
          this.$store.overlay.alert('还差主题名称', '给这套主题起个名字，方便以后一眼认出。');
          return;
        }
        var colors = this.draft.colors;
        var badColor = COLOR_FIELDS.find(function (field) {
          var value = colors[field.key];
          return value !== '' && !HEX_RE.test(value.trim());
        });
        if (badColor) {
          this.$store.overlay.alert(badColor.label + '不是合法色值', '颜色请使用 #RGB 或 #RRGGBB 格式，例如 #1c3f5e，也可以留空。');
          return;
        }
        if (this.draft.enabledAt && Number.isNaN(Date.parse(this.draft.enabledAt))) {
          this.$store.overlay.alert('启用时间不正确', '请选择一个有效的日期和时间，或留空表示立即启用。');
          return;
        }
        var editingId = this.editingId;
        var payload = adminUI.cloneTheme(this.draft);
        payload.name = name;
        payload.id = editingId;
        this.saving = true;
        try {
          var res = editingId
            ? await adminApi.updateTheme(editingId, payload)
            : await adminApi.createTheme(payload);
          if (res.code !== 0) {
            this.$store.overlay.alert('保存失败', adminUI.apiErrorMessage(res));
            return;
          }
          this.$store.toast.success(editingId ? '主题已更新' : '主题已保存');
          var savedId = editingId || (res.data && res.data.id) || '';
          this.dirty = false;
          this.editing = false;
          this.editingId = '';
          this.draft = adminUI.emptyTheme();
          await this.load();
          this.showSavedPreview(savedId || this.overview.currentId);
        } catch (err) {
          this.$store.overlay.alert('保存失败', '网络错误，请稍后再试');
        } finally {
          this.saving = false;
        }
      },
      activateNow: async function (theme) {
        if (this.busyId) return;
        this.busyId = theme.id;
        try {
          var res = await adminApi.activateTheme(theme.id, '');
          if (res.code === 0) {
            this.$store.toast.success('「' + theme.name + '」已启用，网站立即换装');
            await this.load();
            this.showSavedPreview(theme.id);
          } else {
            this.$store.overlay.alert('启用失败', adminUI.apiErrorMessage(res));
          }
        } catch (err) {
          this.$store.overlay.alert('启用失败', '网络错误，请稍后再试');
        } finally {
          this.busyId = '';
        }
      },
      removeTheme: function (theme) {
        var self = this;
        this.$store.overlay.danger('删除「' + theme.name + '」？', '这套主题的配色和布局设置会被移除，网站内容不受影响。', async function () {
          self.busyId = theme.id;
          try {
            var res = await adminApi.deleteTheme(theme.id);
            if (res.code === 0) {
              self.$store.toast.success('主题已删除');
              if (self.editingId === theme.id) self.closeEditor();
              await self.load();
            } else {
              self.$store.overlay.alert('删除失败', adminUI.apiErrorMessage(res));
            }
          } catch (err) {
            self.$store.overlay.alert('删除失败', '网络错误，请稍后再试');
          } finally {
            self.busyId = '';
          }
        });
      },
    };
  });
});
