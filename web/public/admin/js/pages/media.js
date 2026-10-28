document.addEventListener('alpine:init', function () {
  var PAGE_SIZE = 24;
  var VALID_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml'];

  Alpine.data('mediaPage', function () {
    return {
      icon: adminIcon,
      images: [],
      total: 0,
      loading: true,
      error: null,
      keyword: '',
      page: 1,
      dragging: false,
      uploading: false,
      uploadDone: 0,
      uploadTotal: 0,
      lightbox: null,
      init: function () {
        var query = this.$store.app.route.query || {};
        var parsed = parseInt(query.page || '1', 10);
        this.page = Number.isFinite(parsed) && parsed > 0 ? parsed : 1;
        var self = this;
        this.keyHandler = function (e) {
          if (e.key === 'Escape' && self.lightbox) self.lightbox = null;
        };
        document.addEventListener('keydown', this.keyHandler);
        this.load();
      },
      destroy: function () {
        document.removeEventListener('keydown', this.keyHandler);
        this.lightbox = null;
      },
      get filtered() {
        var keyword = this.keyword.trim().toLowerCase();
        if (!keyword) return this.images;
        return this.images.filter(function (img) {
          return img.originalName.toLowerCase().indexOf(keyword) !== -1 || img.fileName.toLowerCase().indexOf(keyword) !== -1;
        });
      },
      get pageCount() { return Math.max(1, Math.ceil(this.filtered.length / PAGE_SIZE)); },
      get currentPage() { return Math.min(this.page, this.pageCount); },
      get pageItems() {
        var start = (this.currentPage - 1) * PAGE_SIZE;
        return this.filtered.slice(start, start + PAGE_SIZE);
      },
      get uploadLabel() {
        return this.uploading ? '上传中 ' + this.uploadDone + '/' + this.uploadTotal : '上传图片';
      },
      get progressWidth() {
        return this.uploadTotal ? (this.uploadDone / this.uploadTotal * 100) + '%' : '0%';
      },
      load: async function () {
        this.loading = true;
        this.error = null;
        try {
          var res = await adminApi.getImages(1, 100);
          if (res.code !== 0) throw new Error(res.message || '无法读取媒体库');
          this.images = res.data || [];
          this.total = res.total || this.images.length;
        } catch (err) {
          this.error = err instanceof Error ? err.message : '网络错误';
        } finally {
          this.loading = false;
        }
      },
      chooseFile: function () {
        var input = this.$refs.file;
        if (input) input.click();
      },
      onFileChange: function (event) {
        var input = event.target;
        if (input.files) this.uploadFiles(input.files);
        input.value = '';
      },
      onDrop: function (event) {
        this.dragging = false;
        if (event.dataTransfer && event.dataTransfer.files) this.uploadFiles(event.dataTransfer.files);
      },
      uploadFiles: async function (files) {
        var self = this;
        var list = Array.prototype.slice.call(files);
        if (list.length === 0 || this.uploading) return;
        this.uploading = true;
        this.uploadDone = 0;
        this.uploadTotal = list.length;
        var uploaded = 0;
        try {
          for (var i = 0; i < list.length; i++) {
            var file = list[i];
            if (VALID_TYPES.indexOf(file.type) === -1) {
              this.$store.toast.warning(file.name + ' 不是支持的图片格式');
            } else if (file.size > 5 * 1024 * 1024) {
              this.$store.toast.warning(file.name + ' 超过 5MB');
            } else {
              try {
                var res = await adminApi.uploadImage(file);
                if (res.code === 0) uploaded++;
                else this.$store.toast.error(file.name + '：' + (res.message || '上传失败'));
              } catch (err) {
                this.$store.toast.error(file.name + ' 上传失败');
              }
            }
            this.uploadDone++;
          }
        } finally {
          this.uploading = false;
        }
        if (uploaded > 0) {
          this.$store.toast.success('已上传 ' + uploaded + ' 张图片');
          this.page = 1;
          await this.load();
        }
      },
      goPage: function (page) {
        if (page < 1 || page > this.pageCount) return;
        this.page = page;
        var url = '/admin/media' + (page > 1 ? '?page=' + page : '');
        history.replaceState(null, '', url);
        window.scrollTo({ top: 0 });
      },
      search: function () {
        this.page = 1;
      },
      clearSearch: function () {
        this.keyword = '';
        this.page = 1;
      },
      copyText: async function (text, success) {
        try {
          await navigator.clipboard.writeText(text);
          this.$store.toast.success(success);
        } catch (err) {
          this.$store.toast.error('复制失败，请手动复制');
        }
      },
      copyImageUrl: function (img) {
        return this.copyText(window.location.origin + img.uploadPath, '图片链接已复制');
      },
      copyMarkdown: function (img) {
        var alt = img.originalName.replace(/\.[^.]+$/, '');
        return this.copyText('![' + alt + '](' + img.uploadPath + ')', 'Markdown 已复制');
      },
      formatSize: function (bytes) {
        return adminUI.formatSize(bytes);
      },
      typeName: function (fileType) {
        return (fileType || '').replace('image/', '');
      },
      remove: function (img) {
        var self = this;
        this.$store.overlay.danger('删除图片', '确定删除「' + img.originalName + '」吗？已被文章引用的图片不能删除。', async function () {
          try {
            var res = await adminApi.deleteImage(img.fileName);
            if (res.code === 0) {
              if (self.lightbox && self.lightbox.fileName === img.fileName) self.lightbox = null;
              self.$store.toast.success('图片已删除');
              await self.load();
            } else if (res.code === 40902) {
              self.$store.overlay.alert('无法删除', res.message || '这张图片仍被内容引用，请先移除引用。');
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
