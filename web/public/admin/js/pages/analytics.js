document.addEventListener('alpine:init', function () {
  Alpine.data('analyticsPage', function () {
    return {
      icon: adminIcon,
      loading: true,
      refreshing: false,
      error: null,
      range: '1d',
      ranges: [
        { key: '5m', label: '5分钟' },
        { key: '1h', label: '1小时' },
        { key: '1d', label: '24小时' },
        { key: '1w', label: '7天' },
        { key: '1m', label: '30天' },
      ],
      kindFilter: 'all',
      kindFilters: [
        { key: 'all', label: '全部' },
        { key: 'product', label: '产品' },
        { key: 'article', label: '文章' },
        { key: 'news', label: '新闻' },
      ],
      kindIcons: { product: 'layers', article: 'article', news: 'send' },
      report: null,
      keyword: '',
      init: function () {
        this.load();
      },
      load: async function (quiet) {
        if (quiet) this.refreshing = true;
        else this.loading = true;
        this.error = null;
        try {
          var res = await adminApi.getAnalytics(this.range);
          if (!res || res.code !== 0) throw new Error((res && res.message) || '无法读取统计数据');
          this.report = res.data || null;
        } catch (err) {
          var message = err instanceof Error ? err.message : '网络错误';
          if (quiet && this.report) {
            Alpine.store('toast').error(message);
          } else {
            this.error = message;
          }
        } finally {
          this.loading = false;
          this.refreshing = false;
        }
      },
      selectRange: function (key) {
        if (this.range === key) return;
        this.range = key;
        this.kindFilter = 'all';
        this.load();
      },
      reload: function () {
        this.load(true);
      },
      get totals() {
        return (this.report && this.report.totals) || { views: 0, users: 0, contents: 0 };
      },
      get kinds() {
        return (this.report && this.report.kinds) || [];
      },
      get series() {
        return (this.report && this.report.series) || [];
      },
      get contents() {
        return (this.report && this.report.contents) || [];
      },
      get rangeLabel() {
        return (this.report && this.report.rangeLabel) || '';
      },
      get bucketLabel() {
        return (this.report && this.report.bucketLabel) || '';
      },
      get headDesc() {
        if (this.totals.views === 0) return '所选时间段还没有访问记录。';
        return '所选时间段共 ' + this.num(this.totals.views) + ' 次访问、' + this.num(this.totals.users) +
          ' 位访客，覆盖 ' + this.num(this.totals.contents) + ' 个内容。';
      },
      get maxSeries() {
        var max = 0;
        this.series.forEach(function (point) {
          max = Math.max(max, point.views, point.users);
        });
        return Math.max(1, max);
      },
      barHeight: function (value) {
        if (!value) return '0%';
        return Math.max(3, Math.round(value / this.maxSeries * 100)) + '%';
      },
      get chartAria() {
        return '访问趋势：' + this.series.map(function (point) {
          return point.hint + ' 访问 ' + point.views + '，访客 ' + point.users;
        }).join('；');
      },
      get filteredContents() {
        var kind = this.kindFilter;
        var keyword = this.keyword.trim().toLowerCase();
        return this.contents.filter(function (item) {
          if (kind !== 'all' && item.kind !== kind) return false;
          if (!keyword) return true;
          return item.title.toLowerCase().indexOf(keyword) !== -1 ||
            item.path.toLowerCase().indexOf(keyword) !== -1;
        });
      },
      kindLabel: function (kind) {
        if (kind === 'product') return '产品';
        if (kind === 'news') return '新闻';
        return '文章';
      },
      kindBadge: function (kind) {
        if (kind === 'product') return 'info';
        if (kind === 'news') return 'warn';
        return 'neutral';
      },
      num: function (value) {
        return (value || 0).toLocaleString('zh-CN');
      },
    };
  });
});
