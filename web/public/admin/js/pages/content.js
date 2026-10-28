document.addEventListener('alpine:init', function () {
  Alpine.data('contentPage', function () {
    return {
      icon: adminIcon,
      items: [],
      dirs: [],
      loading: true,
      error: null,
      tab: 'all',
      keyword: '',
      dirPath: '',
      total: 0,
      articleTotal: 0,
      productTotal: 0,
      pageTotal: 0,
      init: function () {
        var type = this.$store.app.route.query.type;
        this.tab = type === 'page' || type === 'article' || type === 'product' || type === 'draft' ? type : 'all';
        this.dirPath = this.$store.app.route.query.dir || '';
        this.load();
      },
      load: async function () {
        this.loading = true;
        this.error = null;
        try {
          var results = await Promise.all([
            adminApi.getContents(1, 100),
            adminApi.getContents(1, 1, undefined, 'article'),
            adminApi.getContents(1, 1, undefined, 'product'),
            adminApi.getContents(1, 1, undefined, 'page'),
            adminApi.getDirs(),
          ]);
          var all = results[0];
          if (all.code !== 0) {
            this.error = all.message || '无法读取内容列表';
            return;
          }
          this.items = all.data || [];
          this.total = all.total || this.items.length;
          this.articleTotal = results[1].total || 0;
          this.productTotal = results[2].total || 0;
          this.pageTotal = results[3].total || 0;
          this.dirs = (results[4].data || []).slice().sort(function (a, b) {
            return a.navOrder - b.navOrder || a.dirPath.localeCompare(b.dirPath);
          });
        } catch (err) {
          this.error = err instanceof Error ? err.message : '网络错误';
        } finally {
          this.loading = false;
        }
      },
      isPage: function (item) {
        return this.dirs.some(function (d) { return d.dirPath === item.dirPath && d.dirType === 'page'; });
      },
      isProduct: function (item) {
        return this.dirs.some(function (d) { return d.dirPath === item.dirPath && d.dirType === 'product'; });
      },
      typeBadge: function (item) {
        if (this.isProduct(item)) return { tone: 'success', text: '产品' };
        if (this.isPage(item)) return { tone: 'info', text: '页面' };
        return { tone: 'neutral', text: '文章' };
      },
      get draftCount() {
        return this.items.filter(function (item) { return item.draft; }).length;
      },
      get rows() {
        var self = this;
        var list = this.items.slice();
        if (this.tab === 'article') list = list.filter(function (item) { return !self.isPage(item) && !self.isProduct(item); });
        if (this.tab === 'product') list = list.filter(function (item) { return self.isProduct(item); });
        if (this.tab === 'page') list = list.filter(function (item) { return self.isPage(item); });
        if (this.tab === 'draft') list = list.filter(function (item) { return item.draft; });
        if (this.dirPath) list = list.filter(function (item) { return item.dirPath === self.dirPath; });

        var keyword = this.keyword.trim().toLowerCase();
        if (keyword) {
          list = list.filter(function (item) {
            var haystack = [item.title, item.summary || ''].concat(item.tags || []).concat([item.dirPath]).join(' ').toLowerCase();
            return haystack.indexOf(keyword) !== -1;
          });
        }

        return list.sort(function (a, b) {
          var dateDiff = (b.date || '').localeCompare(a.date || '');
          if (dateDiff !== 0) return dateDiff;
          return (a.title || '').localeCompare(b.title || '', 'zh-CN');
        });
      },
      get emptyInfo() {
        if (this.keyword || this.dirPath) {
          return { icon: 'search', title: '没有匹配的内容', desc: '换个关键词或栏目筛选试试。', clear: true };
        }
        if (this.tab === 'draft') return { icon: 'pen', title: '没有草稿', desc: '保存为草稿的内容会显示在这里，发布后才会公开。', href: '/admin/content/new?type=article', action: '写文章' };
        if (this.tab === 'product') return { icon: 'layers', title: '还没有产品', desc: '产品可以设置分类、规格，并关联京东或淘宝下单链接。', href: '/admin/content/new?type=product', action: '上架产品' };
        if (this.tab === 'page') return { icon: 'article', title: '还没有页面', desc: '页面适合放关于我们、公司介绍等固定内容。', href: '/admin/content/new?type=page', action: '创建页面' };
        return { icon: 'article', title: '还没有文章', desc: '文章适合发布新闻、动态和行业观察。', href: '/admin/content/new?type=article', action: '写文章' };
      },
      get showListNote() {
        return this.total > this.items.length && !this.keyword && !this.dirPath;
      },
      editHref: function (item) {
        return '/admin/content/' + encodeURIComponent(item.slug) + '/edit';
      },
      viewHref: function (item) {
        return adminUI.contentHref(item);
      },
      sectionLabel: function (dirPath) {
        return adminUI.sectionLabel(this.dirs, dirPath);
      },
      formatDate: adminUI.formatDate,
      clearFilters: function () {
        this.keyword = '';
        this.dirPath = '';
        this.tab = 'all';
      },
      remove: function (item) {
        var self = this;
        this.$store.overlay.danger('删除内容', '确定删除「' + (item.title || item.slug) + '」吗？删除后无法恢复。', async function () {
          try {
            var res = await adminApi.deleteContent(item.slug);
            if (res.code === 0) {
              self.$store.toast.success('已删除');
              await self.load();
            } else {
              self.$store.overlay.alert('删除失败', res.message || '请稍后再试');
            }
          } catch (err) {
            self.$store.overlay.alert('删除失败', '网络错误，请稍后再试');
          }
        });
      },
    };
  });
});
