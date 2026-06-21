import m from 'mithril';
import { apiService } from '../services/apiService';
import { toastService } from '../components/toast';
import { Icon } from '../icons';
import { topbarState } from '../components/navbar';

interface FeatureCard {
  title: string;
  desc: string;
  icon: string;
  link: string;
}

interface SettingsState {
  form: {
    SITE_TITLE: string;
    SITE_DESC: string;
    HERO_TITLE: string;
    HERO_SUBTITLE: string;
    HERO_IMAGE: string;
    HERO_CTA_TEXT: string;
    HERO_CTA_LINK: string;
    FEATURE_TITLE: string;
    FEATURES: FeatureCard[];
    NEWS_TITLE: string;
    NEWS_LINK: string;
    CONTACT_EMAIL: string;
    CONTACT_PHONE: string;
    CONTACT_ADDRESS: string;
    FOOTER_TEXT: string;
    ICP: string;
  };
  loading: boolean;
  saving: boolean;
}

const emptyForm: SettingsState['form'] = {
  SITE_TITLE: '', SITE_DESC: '',
  HERO_TITLE: '', HERO_SUBTITLE: '', HERO_IMAGE: '',
  HERO_CTA_TEXT: '', HERO_CTA_LINK: '',
  FEATURE_TITLE: '', FEATURES: [],
  NEWS_TITLE: '', NEWS_LINK: '',
  CONTACT_EMAIL: '', CONTACT_PHONE: '', CONTACT_ADDRESS: '',
  FOOTER_TEXT: '', ICP: '',
};

export const SettingsPage: m.Component<{}, SettingsState> = {
  oninit(vnode) {
    Object.assign(vnode.state, { form: { ...emptyForm }, loading: true, saving: false });
    topbarState.title = '站点设置';
    topbarState.crumb = null;
    topbarState.actions = null;
  },

  async oncreate(vnode) {
    try {
      const result = await apiService.getSettings();
      if (result.code === 0 && result.data) {
        const d = result.data as any;
        vnode.state.form = {
          ...emptyForm,
          ...d,
          FEATURES: Array.isArray(d.FEATURES) ? d.FEATURES : [],
        };
      }
    } catch {
      toastService.error('加载设置失败');
    } finally {
      vnode.state.loading = false;
      m.redraw();
    }
  },

  view(vnode) {
    const s = vnode.state;
    if (s.loading) return m('.loading-state', [m('.spinner'), m('p', '加载站点设置...')]);
    const f = s.form;

    return m('.settings-page', [
      m('.page-header', [
        m('div', [
          m('h2', '站点设置'),
          m('.desc', '配置站点基本信息、首页内容与联系方式'),
        ]),
        m('.page-header-actions', [
          m('button.btn.btn-primary', {
            disabled: s.saving,
            onclick: async () => {
              s.saving = true;
              m.redraw();
              try {
                const result = await apiService.updateSettings(f);
                if (result.code === 0) {
                  toastService.success('设置已保存并刷新缓存');
                } else {
                  toastService.error(result.message || '保存失败');
                }
              } catch {
                toastService.error('网络错误');
              } finally {
                s.saving = false;
                m.redraw();
              }
            },
          }, [Icon('save'), s.saving ? '保存中...' : '保存设置']),
        ]),
      ]),

      m('.grid-2', { style: { alignItems: 'start' } }, [
        // Left column
        m('div', [
          // Basic
          m('.card.card-pad.settings-section', [
            m('.section-title', [Icon('globe'), '基本信息']),
            m('.form-row.cols-2', [
              field('站点标题', m('input.form-control', {
                value: f.SITE_TITLE, oninput: (e: Event) => { f.SITE_TITLE = val(e); },
              })),
              field('站点描述', m('input.form-control', {
                value: f.SITE_DESC, oninput: (e: Event) => { f.SITE_DESC = val(e); },
              })),
            ]),
          ]),

          // Hero
          m('.card.card-pad.settings-section', [
            m('.section-title', [Icon('zap'), '首页 Hero 区']),
            m('.form-group', field('Hero 标题', m('input.form-control', {
              value: f.HERO_TITLE, oninput: (e: Event) => { f.HERO_TITLE = val(e); },
            }))),
            m('.form-group', field('Hero 副标题', m('textarea.form-control', {
              value: f.HERO_SUBTITLE, rows: 2, oninput: (e: Event) => { f.HERO_SUBTITLE = val(e); },
            }))),
            m('.form-row.cols-2', [
              field('Hero 背景图 URL', m('input.form-control', {
                value: f.HERO_IMAGE, placeholder: '/uploads/xxx.jpg', oninput: (e: Event) => { f.HERO_IMAGE = val(e); },
              })),
              field('Hero 按钮文字', m('input.form-control', {
                value: f.HERO_CTA_TEXT, oninput: (e: Event) => { f.HERO_CTA_TEXT = val(e); },
              })),
            ]),
            m('.form-group', field('Hero 按钮链接', m('input.form-control', {
              value: f.HERO_CTA_LINK, placeholder: '/products', oninput: (e: Event) => { f.HERO_CTA_LINK = val(e); },
            }))),
          ]),

          // News
          m('.card.card-pad.settings-section', [
            m('.section-title', [Icon('article'), '新闻动态']),
            m('.form-row.cols-2', [
              field('板块标题', m('input.form-control', {
                value: f.NEWS_TITLE, oninput: (e: Event) => { f.NEWS_TITLE = val(e); },
              })),
              field('「更多」链接', m('input.form-control', {
                value: f.NEWS_LINK, placeholder: '/news', oninput: (e: Event) => { f.NEWS_LINK = val(e); },
              })),
            ]),
          ]),
        ]),

        // Right column
        m('div', [
          // Features
          m('.card.card-pad.settings-section', [
            m('.section-title', { style: { justifyContent: 'space-between', display: 'flex' } }, [
              m('span', { style: { display: 'flex', alignItems: 'center', gap: '8px' } }, [Icon('star'), '特色卡片']),
              m('button.btn.btn-ghost.btn-sm', {
                onclick: () => {
                  f.FEATURES.push({ title: '', desc: '', icon: '', link: '' });
                },
              }, [Icon('plus'), '添加']),
            ]),
            field('板块标题', m('input.form-control', {
              value: f.FEATURE_TITLE, oninput: (e: Event) => { f.FEATURE_TITLE = val(e); },
            })),
            m('div', f.FEATURES.map((card, i) =>
              m('.feature-row', [
                m('input.form-control', {
                  value: card.icon, placeholder: '图标', title: '图标 (emoji 或文字)',
                  oninput: (e: Event) => { card.icon = val(e); },
                }),
                m('input.form-control', {
                  value: card.title, placeholder: '标题',
                  oninput: (e: Event) => { card.title = val(e); },
                }),
                m('input.form-control', {
                  value: card.desc, placeholder: '描述',
                  oninput: (e: Event) => { card.desc = val(e); },
                }),
                m('input.form-control', {
                  value: card.link, placeholder: '链接',
                  oninput: (e: Event) => { card.link = val(e); },
                }),
                m('button.btn.btn-danger.btn-icon.btn-sm', {
                  title: '删除',
                  onclick: () => { f.FEATURES.splice(i, 1); },
                }, Icon('trash', { style: { width: '15px', height: '15px' } })),
              ]),
            )),
          ]),

          // Contact
          m('.card.card-pad.settings-section', [
            m('.section-title', [Icon('mail'), '联系方式']),
            m('.form-group', field('邮箱', m('input.form-control', {
              value: f.CONTACT_EMAIL, oninput: (e: Event) => { f.CONTACT_EMAIL = val(e); },
            }))),
            m('.form-group', field('电话', m('input.form-control', {
              value: f.CONTACT_PHONE, oninput: (e: Event) => { f.CONTACT_PHONE = val(e); },
            }))),
            m('.form-group', field('地址', m('input.form-control', {
              value: f.CONTACT_ADDRESS, oninput: (e: Event) => { f.CONTACT_ADDRESS = val(e); },
            }))),
          ]),

          // Footer
          m('.card.card-pad.settings-section', [
            m('.section-title', [Icon('info'), '页脚信息']),
            m('.form-group', field('页脚文字', m('input.form-control', {
              value: f.FOOTER_TEXT, oninput: (e: Event) => { f.FOOTER_TEXT = val(e); },
            }))),
            m('.form-group', field('备案号 (ICP)', m('input.form-control', {
              value: f.ICP, oninput: (e: Event) => { f.ICP = val(e); },
            }))),
          ]),
        ]),
      ]),

      m('div', { style: { display: 'flex', justifyContent: 'flex-end', marginTop: '20px' } }, [
        m('button.btn.btn-primary', {
          disabled: s.saving,
          onclick: async () => {
            s.saving = true;
            m.redraw();
            try {
              const result = await apiService.updateSettings(f);
              if (result.code === 0) {
                toastService.success('设置已保存并刷新缓存');
              } else {
                toastService.error(result.message || '保存失败');
              }
            } catch {
              toastService.error('网络错误');
            } finally {
              s.saving = false;
              m.redraw();
            }
          },
        }, [Icon('save'), s.saving ? '保存中...' : '保存设置']),
      ]),
    ]);
  },
};

function field(label: string, input: m.Vnode): m.Vnode {
  return m('.form-group', [
    m('label.form-label', label),
    input,
  ]);
}

function val(e: Event): string {
  return (e.target as HTMLInputElement | HTMLTextAreaElement).value;
}
