document.addEventListener('alpine:init', function () {
  function toast() { return Alpine.store('toast'); }

  var SIDEBAR_KEY = 'gopaper_sidebar_collapsed';

  Alpine.store('app', {
    route: { name: 'dashboard', path: '/dashboard', params: {}, query: {}, title: '工作台' },
    title: '工作台',
    loggedIn: false,
    sidebarOpen: false,
    sidebarCollapsed: false,
    guard: null,
    init: function () {
      this.loggedIn = adminAuth.isLoggedIn();
      try {
        this.sidebarCollapsed = localStorage.getItem(SIDEBAR_KEY) === '1';
      } catch (err) {
        this.sidebarCollapsed = false;
      }
    },
    isActive: function (prefix) {
      var path = this.route.path;
      return path === prefix || path.indexOf(prefix + '/') === 0;
    },
    get narrowLayout() {
      return window.matchMedia('(max-width: 900px)').matches;
    },
    get sidebarToggleLabel() {
      if (this.narrowLayout) return this.sidebarOpen ? '关闭菜单' : '打开菜单';
      return this.sidebarCollapsed ? '展开侧边栏' : '收起侧边栏';
    },
    get sidebarExpanded() {
      return this.narrowLayout ? this.sidebarOpen : !this.sidebarCollapsed;
    },
    setRoute: function (route) {
      this.route = route;
      this.title = route.title;
      this.loggedIn = adminAuth.isLoggedIn();
    },
    setGuard: function (fn, message) {
      this.guard = { fn: fn, message: message };
    },
    clearGuard: function () {
      this.guard = null;
    },
    toggleSidebar: function () {
      if (window.matchMedia('(max-width: 900px)').matches) {
        this.sidebarOpen = !this.sidebarOpen;
        return;
      }
      this.sidebarCollapsed = !this.sidebarCollapsed;
      try {
        localStorage.setItem(SIDEBAR_KEY, this.sidebarCollapsed ? '1' : '0');
      } catch (err) {
        /* storage unavailable: keep the in-memory state */
      }
    },
    closeSidebar: function () {
      this.sidebarOpen = false;
    },
    logout: function () {
      adminAuth.logout();
      this.loggedIn = false;
      this.sidebarOpen = false;
      if (window.adminNavigate) window.adminNavigate('/login', { replace: true, skipGuard: true });
    },
  });

  Alpine.store('toast', {
    items: [],
    nextId: 0,
    show: function (message, type, duration) {
      var id = this.nextId++;
      this.items.push({ id: id, message: message, type: type });
      setTimeout(function () {
        Alpine.store('toast').items = Alpine.store('toast').items.filter(function (item) { return item.id !== id; });
      }, duration || 3000);
    },
    success: function (message) { this.show(message, 'success'); },
    error: function (message) { this.show(message, 'error'); },
    warning: function (message) { this.show(message, 'warning'); },
  });

  Alpine.store('overlay', {
    open: false,
    type: 'confirm',
    title: '',
    message: '',
    variant: 'warn',
    confirmText: '确定',
    cancelText: '取消',
    promptValue: '',
    onConfirm: null,
    onCancel: null,
    show: function (config) {
      this.type = config.type;
      this.title = config.title;
      this.message = config.message;
      this.variant = config.variant || (config.type === 'confirm' ? 'warn' : 'info');
      this.confirmText = config.confirmText || '确定';
      this.cancelText = config.cancelText || '取消';
      this.promptValue = config.promptDefault || '';
      this.onConfirm = config.onConfirm || null;
      this.onCancel = config.onCancel || null;
      this.open = true;
    },
    close: function () {
      this.open = false;
      this.onConfirm = null;
      this.onCancel = null;
      this.promptValue = '';
    },
    confirm: function (title, message, onConfirm, variant) {
      this.show({ type: 'confirm', title: title, message: message, onConfirm: onConfirm, variant: variant || 'warn' });
    },
    danger: function (title, message, onConfirm) {
      this.show({ type: 'confirm', title: title, message: message, onConfirm: onConfirm, variant: 'danger' });
    },
    alert: function (title, message) {
      this.show({ type: 'alert', title: title, message: message, confirmText: '知道了', variant: 'info' });
    },
    prompt: function (title, message, defaultValue, onSubmit) {
      this.show({
        type: 'prompt', title: title, message: message, promptDefault: defaultValue,
        onConfirm: function () { onSubmit(Alpine.store('overlay').promptValue.trim()); },
      });
    },
    accept: function () {
      var fn = this.onConfirm;
      this.close();
      if (fn) fn();
    },
    cancel: function () {
      var fn = this.onCancel;
      this.close();
      if (fn) fn();
    },
  });

  Alpine.store('picker', {
    open: false,
    images: [],
    loading: false,
    uploading: false,
    keyword: '',
    onPick: null,
    openPicker: function (onPick) {
      this.onPick = onPick;
      this.open = true;
      this.keyword = '';
      if (this.images.length === 0) this.load();
    },
    close: function () {
      this.open = false;
      this.onPick = null;
    },
    load: function () {
      var self = this;
      this.loading = true;
      adminApi.getImages(1, 60).then(function (res) {
        if (res.code === 0) self.images = res.data || [];
        else toast().error(res.message || '图片加载失败');
      }).catch(function () {
        toast().error('图片加载失败');
      }).then(function () {
        self.loading = false;
      });
    },
    filtered: function () {
      var keyword = this.keyword.trim().toLowerCase();
      return this.images.filter(function (img) {
        return !keyword || img.originalName.toLowerCase().indexOf(keyword) !== -1;
      });
    },
    pick: function (image) {
      var onPick = this.onPick;
      this.close();
      if (onPick) onPick(image);
    },
    uploadFiles: function (files) {
      var self = this;
      var list = Array.prototype.slice.call(files);
      if (list.length === 0 || this.uploading) return;
      var valid = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml'];
      var uploaded = 0;
      this.uploading = true;
      var chain = Promise.resolve();
      list.forEach(function (file) {
        chain = chain.then(function () {
          if (valid.indexOf(file.type) === -1) {
            toast().warning(file.name + ' 不是支持的图片格式');
            return null;
          }
          if (file.size > 5 * 1024 * 1024) {
            toast().warning(file.name + ' 超过 5MB');
            return null;
          }
          return adminApi.uploadImage(file).then(function (res) {
            if (res.code === 0) uploaded++;
            else toast().error(file.name + '：' + (res.message || '上传失败'));
          }).catch(function () {
            toast().error(file.name + ' 上传失败');
          });
        });
      });
      chain.then(function () {
        self.uploading = false;
        if (uploaded > 0) {
          toast().success('已上传 ' + uploaded + ' 张图片');
          self.load();
        }
      });
    },
  });

  Alpine.data('adminShell', function () {
    return {
      icon: adminIcon,
      pickFile: function () {
        var input = document.getElementById('picker-file');
        if (input) input.click();
      },
      initShell: function () {
        var self = this;
        this.keyHandler = function (e) {
          if (e.key !== 'Escape') return;
          if (Alpine.store('picker').open) { Alpine.store('picker').close(); return; }
          if (Alpine.store('overlay').open) Alpine.store('overlay').close();
        };
        document.addEventListener('keydown', this.keyHandler);
        this.$watch('$store.overlay.open', function (open) {
          if (!open) return;
          self.$nextTick(function () {
            var dialog = document.querySelector('.overlay-dialog');
            if (!dialog) return;
            var el = dialog.querySelector('input') || dialog.querySelector('.btn-danger') || dialog.querySelector('.btn-primary');
            if (el) {
              el.focus();
              if (el.select) el.select();
            }
          });
        });
      },
    };
  });
});
