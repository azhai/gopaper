document.addEventListener('alpine:init', function () {
  Alpine.data('homepagePage', function () {
    return {
      icon: adminIcon,
      loading: true,
      error: null,
      saving: false,
      dirty: false,
      meta: adminUI.emptySiteMeta(),
      form: adminUI.emptySiteMeta(),
      dirs: [],
      items: [],
      init: function () {
        this.bindLifecycle();
        this.load();
      },
      bindLifecycle: function () {
        var self = this;
        this.$store.app.setGuard(function () { return self.dirty; }, '首页有未保存的修改，离开后会丢失。');
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
      },
      get featured() {
        var self = this;
        return this.items.filter(function (item) {
          return item.position === 'features' && self.dirs.some(function (d) {
            return d.dirPath === item.dirPath && d.dirType === 'page';
          });
        });
      },
      get saveLabel() { return this.saving ? '保存中…' : '保存首页'; },
      get statusText() { return this.dirty ? '首页有未保存的修改' : '所有修改已保存'; },
      load: async function () {
        this.loading = true;
        this.error = null;
        try {
          var results = await Promise.all([
            adminApi.getSite(),
            adminApi.getDirs(),
            adminApi.getContents(1, 100),
          ]);
          var site = results[0];
          if (site.code !== 0 || !site.data) throw new Error(site.message || '无法读取首页设置');
          this.meta = Object.assign(adminUI.emptySiteMeta(), site.data, { FEATURES: site.data.FEATURES || [] });
          this.form = adminUI.cloneSiteMeta(this.meta);
          this.dirs = results[1].data || [];
          this.items = results[2].data || [];
          this.dirty = false;
        } catch (err) {
          this.error = err instanceof Error ? err.message : '网络错误';
        } finally {
          this.loading = false;
        }
      },
      chooseHero: function () {
        var self = this;
        this.$store.picker.openPicker(function (image) {
          self.form.HERO_IMAGE = image.uploadPath;
          self.dirty = true;
        });
      },
      removeHero: function () {
        this.form.HERO_IMAGE = '';
        this.dirty = true;
      },
      addFeature: function () {
        this.form.FEATURES.push({ title: '', desc: '', icon: '', link: '' });
        this.dirty = true;
      },
      removeFeature: function (index) {
        this.form.FEATURES.splice(index, 1);
        this.dirty = true;
      },
      moveFeature: function (index, delta) {
        var list = this.form.FEATURES;
        var target = index + delta;
        if (target < 0 || target >= list.length) return;
        var item = list.splice(index, 1)[0];
        list.splice(target, 0, item);
        this.dirty = true;
      },
      save: async function () {
        if (this.saving) return;
        var invalid = this.form.FEATURES.findIndex(function (card) {
          return !card.title.trim() && (card.desc.trim() || card.link.trim() || card.icon.trim());
        });
        if (invalid >= 0) {
          this.$store.overlay.alert('有一张卡片缺少标题', '第 ' + (invalid + 1) + ' 张卡片填了内容但没有标题，补上标题或删除它。');
          return;
        }
        this.form.FEATURES = this.form.FEATURES.filter(function (card) {
          return card.title.trim() || card.desc.trim() || card.link.trim() || card.icon.trim();
        });
        this.saving = true;
        try {
          var res = await adminApi.saveSite(this.form);
          if (res.code === 0) {
            this.meta = adminUI.cloneSiteMeta(this.form);
            this.dirty = false;
            this.$store.toast.success('首页已更新，前台立即生效');
          } else {
            this.$store.overlay.alert('保存失败', res.message || '请稍后再试');
          }
        } catch (err) {
          this.$store.overlay.alert('保存失败', '网络错误，请稍后再试');
        } finally {
          this.saving = false;
        }
      },
    };
  });
});
