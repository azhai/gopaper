import m from 'mithril';
import { Icon } from '../icons';
let currentOverlay = null;
let promptValue = '';
export const overlayService = {
    show(config) {
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
    setPromptValue: (v) => { promptValue = v; },
};
const variantIcon = {
    info: 'info', warn: 'warning', danger: 'alert',
};
export const Overlay = {
    view() {
        const config = overlayService.getOverlay();
        if (!config)
            return null;
        const variant = config.variant || (config.type === 'confirm' ? 'warn' : 'info');
        return m('.overlay-backdrop', {
            onclick: (e) => {
                if (e.target === e.currentTarget)
                    overlayService.close();
            },
        }, m('.overlay-dialog', [
            m('h3', [
                m(`.icon-circle.${variant}`, Icon(variantIcon[variant])),
                config.title,
            ]),
            m('p', config.message),
            config.type === 'prompt' ? m('input.form-control', {
                value: promptValue,
                oninput: (e) => {
                    promptValue = e.target.value;
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
export function showConfirm(title, message, onConfirm, variant = 'warn') {
    overlayService.show({ type: 'confirm', title, message, onConfirm, variant });
}
export function showDangerConfirm(title, message, onConfirm) {
    overlayService.show({ type: 'confirm', title, message, onConfirm, variant: 'danger' });
}
export function showAlert(title, message) {
    overlayService.show({ type: 'alert', title, message, confirmText: '知道了', variant: 'info' });
}
