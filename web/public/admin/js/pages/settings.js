document.addEventListener('alpine:init', function () {
  var TABS = [
    { value: 'brand', label: '基本信息' },
    { value: 'contact', label: '联系方式' },
    { value: 'footer', label: '页脚' },
    { value: 'sections', label: '栏目' },
    { value: 'advanced', label: '高级' },
  ];

  function normalizeTab(value) {
    var found = TABS.find(function (tab) { return tab.value === value; });
    return found ? found.value : 'brand';
  }

  Alpine.data('settingsPage', function () {
    return {
      icon: adminIcon,
      loading: true,
      error: null,
      tab: 'brand',
      site: adminUI.emptySiteMeta(),
      form: adminUI.emptySiteMeta(),
      siteDirty: false,
      savingSite: false,
      dirs: [],
      sections: [],
      savingSection: null,
      creating: false,
      newSection: { name: '', title: '', type: 'page', navOrder: 10 },
      refreshing: false,
      lastRefresh: '',
      init: function () {
        var query = this.$store.app.route.query || {};
        this.tab = normalizeTab(query.tab);
        this.bindLifecycle();
        this.load();
      },
      bindLifecycle: function () {
        var self = this;
        this.$store.app.setGuard(function () { return self.siteDirty; }, '站点设置还有未保存的修改，离开后会丢失。');
        this.beforeUnload = function (e) {
          if (!self.siteDirty) return;
          e.preventDefault();
          e.returnValue = '';
        };
        window.addEventListener('beforeunload', this.beforeUnload);
      },
      destroy: function () {
        this.$store.app.clearGuard();
        window.removeEventListener('beforeunload', this.beforeUnload);
      },
      get tabs() {
        var self = this;
        return TABS.map(function (tab) {
          return {
            value: tab.value,
            label: tab.label,
            count: tab.value === 'sections' ? self.sections.length : null,
          };
        });
      },
      get isSiteTab() { return this.tab !== 'sections' && this.tab !== 'advanced'; },
      get siteStatus() { return this.siteDirty ? '有未保存的修改' : '所有修改已保存'; },
      get saveLabel() { return this.savingSite ? '保存中…' : '保存设置'; },
      typeLabel: function (type) { return adminUI.typeLabel(type); },
      countFor: function (dirPath) {
        var dir = this.dirs.find(function (item) { return item.dirPath === dirPath; });
        return (dir && dir.articleCount) || 0;
      },
      switchTab: function (value) {
        this.tab = value;
        var route = this.$store.app.route;
        this.$store.app.route = Object.assign({}, route, { query: { tab: value } });
        history.replaceState(null, '', '/admin/settings?tab=' + encodeURIComponent(value));
      },
      load: async function () {
        this.loading = true;
        this.error = null;
        try {
          var results = await Promise.all([adminApi.getSite(), adminApi.getDirs()]);
          var site = results[0];
          if (site.code !== 0 || !site.data) throw new Error(site.message || '无法读取站点设置');
          this.site = Object.assign(adminUI.emptySiteMeta(), site.data, { FEATURES: site.data.FEATURES || [] });
          this.form = adminUI.cloneSiteMeta(this.site);
          this.siteDirty = false;
          this.applyDirs(results[1].data || []);
        } catch (err) {
          this.error = err instanceof Error ? err.message : '网络错误';
        } finally {
          this.loading = false;
        }
      },
      applyDirs: function (dirs) {
        this.dirs = dirs;
        this.sections = dirs
          .filter(function (dir) { return dir.dirPath && dir.dirPath.indexOf('/') === -1; })
          .sort(function (a, b) { return a.navOrder - b.navOrder || a.dirPath.localeCompare(b.dirPath); })
          .map(function (dir) {
            return {
              dirPath: dir.dirPath,
              title: dir.title || dir.dirPath,
              type: dir.dirType || 'page',
              navOrder: dir.navOrder,
              navHide: dir.navHide,
              layout: dir.layout,
              sortOrder: dir.sortOrder,
              dirty: false,
            };
          });
        this.newSection.navOrder = this.nextNavOrder();
      },
      nextNavOrder: function () {
        if (this.sections.length === 0) return 10;
        return Math.max.apply(null, this.sections.map(function (section) { return section.navOrder; })) + 10;
      },
      reloadDirs: async function () {
        var res = await adminApi.getDirs();
        if (res.code === 0) this.applyDirs(res.data || []);
      },
      setField: function (key, value) {
        this.form[key] = value;
        this.siteDirty = true;
      },
      discardSite: function () {
        var self = this;
        if (!this.siteDirty) return;
        this.$store.overlay.confirm('放弃修改？', '本次修改不会保存。', function () {
          self.form = adminUI.cloneSiteMeta(self.site);
          self.siteDirty = false;
        });
      },
      saveSite: async function () {
        if (this.savingSite) return;
        if (!this.form.SITE_TITLE.trim()) {
          this.$store.overlay.alert('还差网站名称', '网站名称会显示在浏览器标题和页头，请填写。');
          return;
        }
        if (this.form.ARTICLES_PER_PAGE < 1 || this.form.ARTICLES_PER_PAGE > 100) {
          this.$store.overlay.alert('每页条数不正确', '列表每页条数请填写 1–100 之间的数字。');
          return;
        }
        var email = this.form.CONTACT_EMAIL.trim();
        if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
          this.$store.overlay.alert('邮箱格式不正确', '请填写完整的邮箱地址，例如 hello@example.com，或留空。');
          return;
        }
        this.savingSite = true;
        try {
          var res = await adminApi.saveSite(this.form);
          if (res.code === 0) {
            this.site = adminUI.cloneSiteMeta(this.form);
            this.siteDirty = false;
            this.$store.toast.success('设置已保存，前台立即生效');
          } else {
            this.$store.overlay.alert('保存失败', res.message || '请稍后再试');
          }
        } catch (err) {
          this.$store.overlay.alert('保存失败', '网络错误，请稍后再试');
        } finally {
          this.savingSite = false;
        }
      },
      createSection: async function () {
        if (this.creating) return;
        var name = this.newSection.name.trim();
        var title = this.newSection.title.trim();
        if (!title) {
          this.$store.overlay.alert('缺少栏目名称', '给新栏目起一个访客看得懂的名字。');
          return;
        }
        if (!/^[a-z0-9-]+$/.test(name)) {
          this.$store.overlay.alert('网址标识不正确', '网址标识只能使用小写字母、数字和连字符，例如 cases。');
          return;
        }
        var exists = this.sections.some(function (section) { return section.dirPath === name; });
        if (exists) {
          this.$store.overlay.alert('网址标识已被使用', '换一个标识，或者直接编辑现有栏目。');
          return;
        }
        this.creating = true;
        try {
          var res = await adminApi.createSection({
            name: name,
            title: title,
            type: this.newSection.type,
            navOrder: this.newSection.navOrder,
            navHide: false,
          });
          if (res.code === 0) {
            this.$store.toast.success('栏目已创建');
            this.newSection = { name: '', title: '', type: 'page', navOrder: this.nextNavOrder() };
            await this.reloadDirs();
          } else {
            this.$store.overlay.alert('创建失败', adminUI.apiErrorMessage(res));
          }
        } catch (err) {
          this.$store.overlay.alert('创建失败', '网络错误，请稍后再试');
        } finally {
          this.creating = false;
        }
      },
      resetSection: function (section) {
        var original = this.dirs.find(function (dir) { return dir.dirPath === section.dirPath; });
        if (!original) return;
        section.title = original.title || original.dirPath;
        section.type = original.dirType || 'page';
        section.navOrder = original.navOrder;
        section.navHide = original.navHide;
        section.dirty = false;
      },
      saveSection: async function (section) {
        if (this.savingSection) return;
        if (!section.title.trim()) {
          this.$store.overlay.alert('栏目名称不能为空', '给栏目起一个名字再保存。');
          return;
        }
        this.savingSection = section.dirPath;
        try {
          var res = await adminApi.updateSection({
            dirPath: section.dirPath,
            title: section.title.trim(),
            type: section.type,
            navOrder: section.navOrder,
            navHide: section.navHide,
            layout: section.layout,
            sortOrder: section.sortOrder,
          });
          if (res.code === 0) {
            this.$store.toast.success('栏目已更新');
            await this.reloadDirs();
          } else {
            this.$store.overlay.alert('保存失败', adminUI.apiErrorMessage(res));
          }
        } catch (err) {
          this.$store.overlay.alert('保存失败', '网络错误，请稍后再试');
        } finally {
          this.savingSection = null;
        }
      },
      refreshCache: async function () {
        this.refreshing = true;
        try {
          var res = await adminApi.refreshCache();
          if (res.code === 0 && res.data) {
            this.lastRefresh = '已重建索引：' + res.data.articleCount + ' 条内容，用时 ' + res.data.duration + '。';
            this.$store.toast.success('内容索引已刷新');
          } else {
            this.$store.overlay.alert('刷新失败', res.message || '请稍后再试');
          }
        } catch (err) {
          this.$store.overlay.alert('刷新失败', '网络错误，请稍后再试');
        } finally {
          this.refreshing = false;
        }
      },
    };
  });
});
