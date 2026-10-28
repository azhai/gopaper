document.addEventListener('alpine:init', function () {
  Alpine.data('editorPage', function () {
    return {
      icon: adminIcon,
      loading: true,
      loadError: null,
      isEdit: false,
      isPage: false,
      isProduct: false,
      slug: '',
      dirs: [],
      form: {
        dirPath: '', title: '', slug: '', author: '', date: adminUI.today(), tags: '', summary: '',
        weight: 0, position: '', comments: true, draft: true, category: '', subCategory: '',
        cover: '', specs: [], improvements: [], content: '',
      },
      tab: 'write',
      previewHtml: '',
      previewLoading: false,
      saving: null,
      dirty: false,
      showAdvanced: false,
      lastSaved: '',
      emptyPreview: '<p class="muted">正文还是空的，切回编写开始输入。</p>',
      init: function () {
        var route = this.$store.app.route;
        this.slug = (route.params.slug || '').trim();
        this.isEdit = !!this.slug;
        this.isPage = route.query.type === 'page';
        this.isProduct = route.query.type === 'product';
        this.bindLifecycle();
        this.load();
      },
      bindLifecycle: function () {
        var self = this;
        this.$store.app.setGuard(function () { return self.dirty; }, '这篇内容有未保存的修改，离开后会丢失。');
        this.beforeUnload = function (e) {
          if (!self.dirty) return;
          e.preventDefault();
          e.returnValue = '';
        };
        this.keyHandler = function (e) {
          if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 's') {
            e.preventDefault();
            self.save('draft');
          }
        };
        window.addEventListener('beforeunload', this.beforeUnload);
        window.addEventListener('keydown', this.keyHandler);
      },
      destroy: function () {
        this.$store.app.clearGuard();
        window.removeEventListener('beforeunload', this.beforeUnload);
        window.removeEventListener('keydown', this.keyHandler);
      },
      get isArticle() { return !this.isPage && !this.isProduct; },
      get noun() { return this.isProduct ? '产品' : this.isPage ? '页面' : '文章'; },
      get availableDirs() {
        if (this.isProduct) return this.dirs.filter(function (d) { return d.dirType === 'product'; });
        if (this.isPage) return this.dirs.filter(function (d) { return d.dirType === 'page'; });
        return this.dirs.filter(function (d) { return d.dirType !== 'page' && d.dirType !== 'product'; });
      },
      get previewable() { return this.isEdit && !this.form.draft; },
      get publicHref() { return adminUI.contentHref(this.form); },
      get titlePlaceholder() {
        return this.isProduct ? '产品名称' : this.isPage ? '给页面起个名字' : '标题';
      },
      get contentPlaceholder() {
        if (this.isProduct) return '介绍产品亮点、适用场景和使用说明。支持 Markdown。';
        if (this.isPage) return '在这里写页面正文。支持 Markdown，可插入图片、标题和列表。';
        return '从这里开始写正文。支持 Markdown，可插入图片、标题和列表。';
      },
      get pageLabel() { return adminUI.sectionLabel(this.dirs, this.form.dirPath); },
      get charCount() { return adminUI.textLength(this.form.content); },
      get footStatus() { return this.dirty ? '有未保存的修改' : (this.lastSaved || ''); },
      get publishHint() { return this.form.draft ? '草稿不会出现在网站上' : '保存后前台立即可见'; },
      load: async function () {
        this.loading = true;
        this.loadError = null;
        try {
          var dirsRes = await adminApi.getDirs();
          if (dirsRes.code !== 0) throw new Error(dirsRes.message || '栏目加载失败');
          this.dirs = (dirsRes.data || []).slice().sort(function (a, b) {
            return a.navOrder - b.navOrder || a.dirPath.localeCompare(b.dirPath);
          });
          if (this.isEdit) {
            var res = await adminApi.getContent(this.slug);
            if (res.code !== 0 || !res.data) throw new Error(res.message || '内容不存在或已被删除');
            this.fillForm(res.data);
          } else {
            var dirs = this.availableDirs;
            this.form.dirPath = dirs.length ? dirs[0].dirPath : '';
            if (dirs.length === 0) this.showAdvanced = true;
          }
          this.dirty = false;
        } catch (err) {
          this.loadError = err instanceof Error ? err.message : '网络错误';
        } finally {
          this.loading = false;
        }
      },
      fillForm: function (data) {
        var self = this;
        this.isPage = this.dirs.some(function (d) { return d.dirPath === data.dirPath && d.dirType === 'page'; });
        this.isProduct = this.dirs.some(function (d) { return d.dirPath === data.dirPath && d.dirType === 'product'; });
        var category = adminUI.splitCategory(data.category || '');
        this.form = {
          dirPath: data.dirPath || '',
          title: data.title || '',
          slug: data.slug || '',
          author: data.author || '',
          date: data.date || adminUI.today(),
          tags: (data.tags || []).join(', '),
          summary: data.summary || '',
          weight: data.weight || 0,
          position: data.position || '',
          comments: data.comments !== false,
          draft: !!data.draft,
          category: category[0],
          subCategory: category[1],
          cover: data.cover || '',
          specs: (data.specs || []).map(function (spec) { return Object.assign({}, spec); }),
          improvements: (data.improvements || []).map(function (item) { return Object.assign({}, item); }),
          content: data.content || '',
        };
        this.slug = data.slug;
        this.lastSaved = ('最近更新 ' + (data.date || '')).trim();
        this.previewHtml = '';
        if (window.highlightCode) this.$nextTick(function () { if (self.$refs.preview) window.highlightCode(self.$refs.preview); });
      },
      insertMarkdown: function (before, after, placeholder) {
        var self = this;
        var textarea = this.$refs.content;
        var content = this.form.content;
        var start = textarea ? textarea.selectionStart : content.length;
        var end = textarea ? textarea.selectionEnd : content.length;
        var selected = content.slice(start, end) || placeholder;
        this.form.content = content.slice(0, start) + before + selected + after + content.slice(end);
        this.tab = 'write';
        this.dirty = true;
        var nextStart = start + before.length;
        var nextEnd = nextStart + selected.length;
        this.$nextTick(function () {
          var el = self.$refs.content;
          if (!el) return;
          el.focus();
          el.setSelectionRange(nextStart, nextEnd);
        });
      },
      insert: function (kind) {
        var map = {
          heading: ['## ', '', '小标题'],
          bold: ['**', '**', '重点文字'],
          italic: ['*', '*', '文字'],
          list: ['- ', '', '列表项'],
          quote: ['> ', '', '引用内容'],
          code: ['```\n', '\n```', '代码'],
        };
        var pair = map[kind];
        if (pair) this.insertMarkdown(pair[0], pair[1], pair[2]);
      },
      insertLink: function () {
        var self = this;
        this.$store.overlay.prompt('插入链接', '输入要跳转的网址', 'https://', function (url) {
          if (!url) return;
          self.insertMarkdown('[', '](' + url + ')', '链接文字');
        });
      },
      pickImage: async function () {
        var self = this;
        this.$store.picker.openPicker(function (image) {
          var alt = image.originalName.replace(/\.[^.]+$/, '');
          self.insertMarkdown('![' + alt + '](', ')', image.uploadPath);
        });
      },
      chooseCover: function () {
        var self = this;
        this.$store.picker.openPicker(function (image) {
          self.form.cover = image.uploadPath;
          self.dirty = true;
        });
      },
      openPreview: async function () {
        var self = this;
        this.tab = 'preview';
        this.previewLoading = true;
        try {
          var res = await adminApi.previewMarkdown(this.form.content);
          this.previewHtml = res.code === 0 && res.data ? res.data.html : '<p class="muted">预览渲染失败。</p>';
        } catch (err) {
          this.previewHtml = '<p class="muted">预览渲染失败。</p>';
        } finally {
          this.previewLoading = false;
          this.$nextTick(function () {
            if (window.highlightCode && self.$refs.preview) window.highlightCode(self.$refs.preview);
          });
        }
      },
      save: function (mode) {
        if (this.saving) return false;
        if (!this.form.title.trim()) {
          this.$store.overlay.alert('还差一个标题', '给内容起个标题再保存。');
          return false;
        }
        if (!this.form.dirPath) {
          this.$store.overlay.alert('请选择栏目', '内容需要放在一个栏目下。可以到「站点设置 → 栏目」新建。');
          return false;
        }
        if (this.isProduct) {
          if (this.form.specs.length === 0) {
            this.$store.overlay.alert('还没有规格', '产品至少需要一个规格，每个规格都要填写 SKU。');
            return false;
          }
          var invalid = this.form.specs.findIndex(function (spec) { return !spec.name.trim() || !spec.sku || !spec.sku.trim(); });
          if (invalid >= 0) {
            this.$store.overlay.alert('规格信息不完整', '规格 ' + (invalid + 1) + ' 需要填写名称和 SKU。');
            return false;
          }
          var invalidImprovement = this.form.improvements.findIndex(function (item) { return !item.date || !item.content.trim(); });
          if (invalidImprovement >= 0) {
            this.$store.overlay.alert('听劝记录不完整', '记录 ' + (invalidImprovement + 1) + ' 需要填写日期和改进内容。');
            return false;
          }
        }
        if (mode === 'publish' && !this.form.content.trim()) {
          var self = this;
          this.$store.overlay.confirm('正文还是空的', '确定要发布一篇没有正文的内容吗？', function () {
            self.persist(mode);
          });
          return false;
        }
        this.persist(mode);
        return true;
      },
      persist: async function (mode) {
        var self = this;
        this.saving = mode;
        var isArticle = this.isArticle;
        var payload = {
          dirPath: this.form.dirPath,
          title: this.form.title.trim(),
          slug: this.form.slug.trim() || undefined,
          author: isArticle ? this.form.author.trim() : '',
          date: isArticle ? this.form.date : '',
          tags: isArticle ? this.tagsFrom(this.form.tags) : [],
          comments: isArticle ? this.form.comments : true,
          weight: Number(this.form.weight) || 0,
          position: this.isPage ? this.form.position : '',
          draft: mode === 'draft',
          summary: this.form.summary.trim(),
          category: this.isProduct ? this.joinCategory(this.form.category, this.form.subCategory) : '',
          cover: this.isProduct ? this.form.cover : '',
          specs: this.isProduct ? this.form.specs.map(function (spec) {
            return {
              name: spec.name.trim(),
              price: spec.price ? spec.price.trim() : '',
              sku: spec.sku ? spec.sku.trim() : '',
              jdUrl: spec.jdUrl ? spec.jdUrl.trim() : '',
              taobaoUrl: spec.taobaoUrl ? spec.taobaoUrl.trim() : '',
            };
          }) : [],
          improvements: this.isProduct ? this.form.improvements.map(function (item) {
            return { date: item.date, content: item.content.trim(), visible: item.visible };
          }) : [],
          content: this.form.content,
        };
        try {
          if (this.isEdit) {
            var res = await adminApi.updateContent(this.slug, payload);
            if (res.code !== 0) {
              this.$store.overlay.alert('保存失败', adminUI.apiErrorMessage(res));
              return;
            }
            this.form.draft = mode === 'draft';
            this.dirty = false;
            this.lastSaved = '已保存 ' + new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
            this.$store.toast.success(mode === 'draft' ? '已保存为草稿' : '已更新，前台立即生效');
          } else {
            var created = await adminApi.createContent(payload);
            if (created.code !== 0 || !created.data || !created.data.slug) {
              this.$store.overlay.alert('保存失败', adminUI.apiErrorMessage(created));
              return;
            }
            this.dirty = false;
            this.$store.toast.success(mode === 'draft' ? '已保存为草稿' : '已发布，前台立即生效');
            window.adminNavigate('/content/' + encodeURIComponent(created.data.slug) + '/edit', { replace: true, skipGuard: true });
          }
        } catch (err) {
          this.$store.overlay.alert('保存失败', err instanceof Error ? err.message : '网络错误，请稍后再试');
        } finally {
          this.saving = null;
        }
      },
      tagsFrom: function (value) {
        return value.split(/[,，]/).map(function (tag) { return tag.trim(); }).filter(Boolean);
      },
      joinCategory: function (parent, child) {
        return [parent.trim(), child.trim()].filter(Boolean).join('/');
      },
      addSpec: function () {
        this.form.specs.push({ name: '', price: '', sku: '', jdUrl: '', taobaoUrl: '' });
        this.dirty = true;
      },
      removeSpec: function (index) {
        this.form.specs.splice(index, 1);
        this.dirty = true;
      },
      addImprovement: function () {
        this.form.improvements.push({ date: adminUI.today(), content: '', visible: true });
        this.dirty = true;
      },
      removeImprovement: function (index) {
        this.form.improvements.splice(index, 1);
        this.dirty = true;
      },
      chooseCoverFile: function () {
        this.chooseCover();
      },
      removeCover: function () {
        this.form.cover = '';
        this.dirty = true;
      },
      leave: function () {
        var self = this;
        if (!this.dirty) {
          window.adminNavigate('/content', { skipGuard: true });
          return;
        }
        this.$store.overlay.confirm('放弃未保存的修改？', '离开后本次编辑的内容不会保留。', function () {
          self.dirty = false;
          self.$store.app.clearGuard();
          window.adminNavigate('/content', { skipGuard: true });
        });
      },
      removeContent: function () {
        var self = this;
        this.$store.overlay.danger('删除内容', '确定删除「' + (this.form.title || this.slug) + '」吗？删除后无法恢复。', async function () {
          try {
            var res = await adminApi.deleteContent(self.slug);
            if (res.code === 0) {
              self.dirty = false;
              self.$store.app.clearGuard();
              self.$store.toast.success('已删除');
              window.adminNavigate('/content', { skipGuard: true });
            } else {
              self.$store.overlay.alert('删除失败', res.message || '请稍后再试');
            }
          } catch (err) {
            self.$store.overlay.alert('删除失败', '网络错误，请稍后再试');
          }
        });
      },
      toDraft: function () {
        var self = this;
        this.$store.overlay.confirm('转为草稿？', '转为草稿后，这个页面会从前台隐藏。', function () {
          self.persist('draft');
        });
      },
    };
  });
});
