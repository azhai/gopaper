import m from 'mithril';
import { apiService, type LayoutConfig, type TemplateInfo, type Region } from '../services/apiService';
import { toastService } from '../components/toast';
import { Icon } from '../icons';
import { topbarState } from '../components/navbar';

interface LayoutState {
  config: LayoutConfig;
  loading: boolean;
  saving: boolean;
  expandedIdx: number | null;
}

const emptyConfig: LayoutConfig = { templates: [] };

export const LayoutPage: m.Component<{}, LayoutState> = {
  oninit(vnode) {
    Object.assign(vnode.state, { config: { ...emptyConfig }, loading: true, saving: false, expandedIdx: null });
    topbarState.title = '布局管理';
    topbarState.crumb = null;
    topbarState.actions = null;
  },

  async oncreate(vnode) {
    try {
      const result = await apiService.getLayouts();
      if (result.code === 0 && result.data) {
        vnode.state.config = result.data;
      }
    } catch {
      toastService.error('加载布局配置失败');
    } finally {
      vnode.state.loading = false;
      m.redraw();
    }
  },

  view(vnode) {
    const s = vnode.state;
    if (s.loading) return m('.loading-state', [m('.spinner'), m('p', '加载布局配置...')]);
    const cfg = s.config;

    return m('.layout-page', [
      m('.page-header', [
        m('div', [
          m('h2', '布局管理'),
          m('.desc', '管理页面模板和区域定义，页面可分配到不同区域'),
        ]),
        m('.page-header-actions', [
          m('button.btn.btn-ghost.btn-sm', {
            onclick: () => addTemplate(s),
          }, [Icon('plus'), '新增模板']),
          m('button.btn.btn-primary', {
            disabled: s.saving,
            onclick: async () => {
              s.saving = true;
              m.redraw();
              try {
                const result = await apiService.saveLayouts(cfg);
                if (result.code === 0) {
                  toastService.success('布局配置已保存');
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
          }, [Icon('save'), s.saving ? '保存中...' : '保存配置']),
        ]),
      ]),

      m('.templates-list', cfg.templates.map((tmpl, i) =>
        m('.card.template-card', [
          m('.template-header', {
            onclick: () => { s.expandedIdx = s.expandedIdx === i ? null : i; },
          }, [
            m('.template-info', [
              m('.template-name', tmpl.name || '(未命名)'),
              m('.template-meta', [
                tmpl.title,
                tmpl.file ? ` · ${tmpl.file}` : '',
                tmpl.desc ? ` · ${tmpl.desc}` : '',
              ]),
            ]),
            m('.template-actions', [
              m('span.region-count', `${tmpl.regions.length} 个区域`),
              Icon(s.expandedIdx === i ? 'chevron-down' : 'chevron-right'),
            ]),
          ]),

          s.expandedIdx === i ? m('.template-body', [
            m('.form-row.cols-3', [
              field('模板标识', m('input.form-control', {
                value: tmpl.name, placeholder: 'home',
                oninput: (e: Event) => { tmpl.name = val(e); },
              })),
              field('显示名称', m('input.form-control', {
                value: tmpl.title, placeholder: '首页模板',
                oninput: (e: Event) => { tmpl.title = val(e); },
              })),
              field('模板文件', m('input.form-control', {
                value: tmpl.file, placeholder: 'index.html',
                oninput: (e: Event) => { tmpl.file = val(e); },
              })),
            ]),
            m('.form-group', field('描述', m('input.form-control', {
              value: tmpl.desc, placeholder: '模板描述',
              oninput: (e: Event) => { tmpl.desc = val(e); },
            }))),

            m('.regions-section', [
              m('.regions-header', [
                m('span', [Icon('layout'), '区域定义']),
                m('button.btn.btn-ghost.btn-sm', {
                  onclick: () => {
                    tmpl.regions.push({ name: '', title: '' });
                  },
                }, [Icon('plus'), '添加区域']),
              ]),
              m('.regions-list', tmpl.regions.map((region, j) =>
                m('.region-row', [
                  m('input.form-control', {
                    value: region.name, placeholder: '区域标识 (如 hero)',
                    oninput: (e: Event) => { region.name = val(e); },
                  }),
                  m('input.form-control', {
                    value: region.title, placeholder: '区域名称 (如 横幅区)',
                    oninput: (e: Event) => { region.title = val(e); },
                  }),
                  m('button.btn.btn-danger.btn-icon.btn-sm', {
                    title: '删除区域',
                    onclick: () => { tmpl.regions.splice(j, 1); },
                  }, Icon('trash', { style: { width: '15px', height: '15px' } })),
                ]),
              )),
            ]),

            m('.template-remove', [
              m('button.btn.btn-danger.btn-sm', {
                onclick: () => {
                  if (confirm(`确认删除模板 "${tmpl.name}"？`)) {
                    cfg.templates.splice(i, 1);
                    s.expandedIdx = null;
                  }
                },
              }, [Icon('trash'), '删除模板']),
            ]),
          ]) : null,
        ]),
      )),
    ]);
  },
};

function addTemplate(s: LayoutState) {
  s.config.templates.push({
    name: '',
    title: '',
    file: '',
    desc: '',
    regions: [],
  });
  s.expandedIdx = s.config.templates.length - 1;
}

function field(label: string, input: m.Vnode): m.Vnode {
  return m('.form-group', [
    m('label.form-label', label),
    input,
  ]);
}

function val(e: Event): string {
  return (e.target as HTMLInputElement).value;
}
