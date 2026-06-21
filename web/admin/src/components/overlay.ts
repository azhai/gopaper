import m from 'mithril';
import { Icon, IconName } from '../icons';

export interface OverlayConfig {
  type: 'confirm' | 'alert' | 'prompt';
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'info' | 'warn' | 'danger';
  onConfirm?: () => void;
  onCancel?: () => void;
  promptDefault?: string;
}

let currentOverlay: OverlayConfig | null = null;
let promptValue = '';

export const overlayService = {
  show(config: OverlayConfig) {
    currentOverlay = config;
    promptValue = config.promptDefault || '';
    m.redraw();
  },

  close() {
    currentOverlay = null;
    promptValue = '';
    m.redraw();
  },

  getOverlay: () => currentOverlay,
  getPromptValue: () => promptValue,
  setPromptValue: (v: string) => { promptValue = v; },
};

const variantIcon: Record<string, IconName> = {
  info: 'info', warn: 'warning', danger: 'alert',
};

export const Overlay: m.Component = {
  view() {
    const config = overlayService.getOverlay();
    if (!config) return null;

    const variant = config.variant || (config.type === 'confirm' ? 'warn' : 'info');

    return m('.overlay-backdrop', {
      onclick: (e: Event) => {
        if (e.target === e.currentTarget) overlayService.close();
      },
    }, m('.overlay-dialog', [
      m('h3', [
        m(`.icon-circle.${variant}`, Icon(variantIcon[variant])),
        config.title,
      ]),
      m('p', config.message),
      config.type === 'prompt' ? m('input.form-control', {
        value: promptValue,
        oninput: (e: Event) => {
          promptValue = (e.target as HTMLInputElement).value;
        },
      }) : null,
      m('.overlay-actions', [
        config.type === 'confirm' ? m('button.btn.btn-ghost', {
          onclick: () => {
            overlayService.close();
            config.onCancel?.();
          },
        }, config.cancelText || '取消') : null,
        m('button.btn', {
          class: variant === 'danger' ? 'btn-danger' : 'btn-primary',
          onclick: () => {
            overlayService.close();
            config.onConfirm?.();
          },
        }, config.confirmText || '确定'),
      ]),
    ]));
  },
};

export function showConfirm(title: string, message: string, onConfirm: () => void, variant: 'warn' | 'danger' = 'warn') {
  overlayService.show({ type: 'confirm', title, message, onConfirm, variant });
}

export function showDangerConfirm(title: string, message: string, onConfirm: () => void) {
  overlayService.show({ type: 'confirm', title, message, onConfirm, variant: 'danger' });
}

export function showAlert(title: string, message: string) {
  overlayService.show({ type: 'alert', title, message, confirmText: '知道了', variant: 'info' });
}
