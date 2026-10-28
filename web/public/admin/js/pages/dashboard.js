document.addEventListener('alpine:init', function () {
  Alpine.data('dashboardPage', function () {
    return {
      icon: adminIcon,
      loading: true,
      error: null,
      site: null,
      dirs: [],
      items: [],
      total: 0,
      articleTotal: 0,
      pageTotal: 0,
      productTotal: 0,
      imageTotal: 0,
      init: function () {
        this.load();
      },
      load: async function () {
        this.loading = true;
        this.error = null;
        try {
          var results = await Promise.all([
            adminApi.getSite(),
            adminApi.getDirs(),
            adminApi.getContents(1, 100),
            adminApi.getContents(1, 1, undefined, 'article'),
            adminApi.getContents(1, 1, undefined, 'page'),
            adminApi.getContents(1, 1, undefined, 'product'),
            adminApi.getImages(1, 1),
          ]);
          var contents = results[2];
          if (contents.code !== 0) throw new Error(contents.message || '无法读取内容');
          this.site = results[0].code === 0 && results[0].data ? results[0].data : null;
          this.dirs = (results[1].data || []).slice().sort(function (a, b) {
            return a.navOrder - b.navOrder || a.dirPath.localeCompare(b.dirPath);
          });
          this.items = (contents.data || []).slice().sort(function (a, b) {
            return (b.date || '').localeCompare(a.date || '');
          });
          this.total = contents.total || this.items.length;
          this.articleTotal = results[3].total || 0;
          this.pageTotal = results[4].total || 0;
          this.productTotal = results[5].total || 0;
          this.imageTotal = results[6].total || 0;
        } catch (err) {
          this.error = err instanceof Error ? err.message : '网络错误';
        } finally {
          this.loading = false;
        }
      },
      get greeting() {
        var hour = new Date().getHours();
        if (hour < 6) return '夜深了，注意休息';
        if (hour < 12) return '早上好';
        if (hour < 14) return '中午好';
        if (hour < 18) return '下午好';
        return '晚上好';
      },
      get todayLabel() {
        return new Date().toLocaleDateString('zh-CN', {
          year: 'numeric', month: 'long', day: 'numeric', weekday: 'long',
        });
      },
      get pageDesc() {
        return this.todayLabel + ' · 共 ' + this.total + ' 条内容（文章 ' + this.articleTotal +
          ' · 产品 ' + this.productTotal + ' · 页面 ' + this.pageTotal + '），' +
          this.drafts.length + ' 条草稿待处理。';
      },
      get drafts() {
        return this.items.filter(function (item) { return item.draft; });
      },
      get recent() {
        return this.items.slice(0, 6);
      },
      get health() {
        var site = this.site;
        var published = this.items.some(function (item) { return !item.draft; });
        var featured = (site && site.FEATURES ? site.FEATURES.length : 0) > 0 ||
          this.items.some(function (item) { return item.position === 'features'; });
        return [
          { label: '设置网站名称', done: !!(site && site.SITE_TITLE.trim()), hint: '显示在浏览器标题和网站页头', href: '/settings?tab=brand', action: '去设置' },
          { label: '填写首页大标题', done: !!(site && site.HERO_TITLE.trim()), hint: '访客打开网站看到的第一句话', href: '/homepage', action: '去装修' },
          { label: '添加联系方式', done: !!(site && (site.CONTACT_EMAIL.trim() || site.CONTACT_PHONE.trim())), hint: '让客户能直接联系你', href: '/settings?tab=contact', action: '去填写' },
          { label: '创建至少一个栏目', done: this.dirs.length > 0, hint: '栏目决定网站导航和内容归属', href: '/settings?tab=sections', action: '去创建' },
          { label: '发布第一篇内容', done: published, hint: '草稿不会出现在网站上', href: '/content/new?type=article', action: '去发布' },
          { label: '添加服务卡片', done: featured, hint: '在首页展示你能提供的产品与服务', href: '/homepage', action: '去添加' },
        ];
      },
      get doneCount() {
        return this.health.filter(function (item) { return item.done; }).length;
      },
      get progressWidth() {
        return Math.round(this.doneCount / this.health.length * 100) + '%';
      },
      editHref: function (item) {
        return '/admin/content/' + encodeURIComponent(item.slug) + '/edit';
      },
      viewHref: function (item) {
        return adminUI.contentHref(item);
      },
      sectionHref: function (dir) {
        return '/admin/content?dir=' + encodeURIComponent(dir.dirPath);
      },
      sectionLabel: function (dirPath) {
        return adminUI.sectionLabel(this.dirs, dirPath);
      },
      formatDate: adminUI.formatDate,
      typeLabel: adminUI.typeLabel,
    };
  });
});
