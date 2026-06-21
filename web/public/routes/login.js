import m from 'mithril';
import { apiService, authService } from '../services/apiService';
import { showAlert } from '../components/overlay';
import { Icon } from '../icons';
export const LoginPage = {
    view() {
        return m('.login-page', m('.login-card', [
            m('.login-logo', [
                m('.login-logo-icon', Icon('gopaper')),
            ]),
            m('.login-title', 'GoPaper'),
            m('.login-subtitle', '内容管理后台 · 请登录以继续'),
            m('form', {
                onsubmit: async (e) => {
                    e.preventDefault();
                    const form = e.target;
                    const username = form.querySelector('[name="username"]').value;
                    const password = form.querySelector('[name="password"]').value;
                    try {
                        const result = await apiService.login(username, password);
                        if (result.code === 0 && result.data) {
                            authService.setToken(result.data.token);
                            m.route.set('/dashboard');
                        }
                        else {
                            showAlert('登录失败', result.message || '用户名或密码错误');
                        }
                    }
                    catch {
                        showAlert('登录失败', '网络错误，请稍后重试');
                    }
                },
            }, [
                m('.form-group', [
                    m('label.form-label', '用户名'),
                    m('input.form-control', {
                        name: 'username', type: 'text', required: true, placeholder: '请输入用户名',
                        autocomplete: 'username',
                    }),
                ]),
                m('.form-group', [
                    m('label.form-label', '密码'),
                    m('input.form-control', {
                        name: 'password', type: 'password', required: true, placeholder: '请输入密码',
                        autocomplete: 'current-password',
                    }),
                ]),
                m('button.btn.btn-primary', {
                    type: 'submit',
                    style: { width: '100%', padding: '12px', fontSize: '15px', marginTop: '6px' },
                }, [Icon('logout', { style: { transform: 'rotate(180deg)' } }), '登 录']),
            ]),
        ]));
    },
};
