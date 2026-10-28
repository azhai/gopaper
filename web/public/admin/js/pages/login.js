document.addEventListener('alpine:init', function () {
  Alpine.data('loginPage', function () {
    return {
      username: '',
      password: '',
      error: '',
      submitting: false,
      showPassword: false,
      init: function () {
        if (adminAuth.isLoggedIn()) {
          Alpine.store('app').loggedIn = true;
          window.adminNavigate('/dashboard', { replace: true, skipGuard: true });
        }
      },
      submit: async function () {
        if (this.submitting) return;
        var username = this.username.trim();
        if (!username || !this.password) {
          this.error = '请填写用户名和密码';
          return;
        }
        this.submitting = true;
        this.error = '';
        try {
          var res = await adminApi.login(username, this.password);
          if (res.code === 0 && res.data && res.data.token) {
            adminAuth.setToken(res.data.token);
            Alpine.store('app').loggedIn = true;
            window.adminNavigate('/dashboard', { replace: true, skipGuard: true });
            return;
          }
          this.error = res.message || '用户名或密码错误';
        } catch (err) {
          this.error = err instanceof Error ? err.message : '网络错误，请稍后再试';
        } finally {
          this.submitting = false;
        }
      },
    };
  });
});
