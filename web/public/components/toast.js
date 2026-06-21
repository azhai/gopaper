import m from 'mithril';
import { Icon } from '../icons';
let toasts = [];
let nextId = 0;
export const toastService = {
    show(config) {
        const id = nextId++;
        toasts.push({ ...config, id });
        m.redraw();
        setTimeout(() => {
            toasts = toasts.filter(t => t.id !== id);
            m.redraw();
        }, config.duration || 3000);
    },
    success: (message) => toastService.show({ message, type: 'success' }),
    error: (message) => toastService.show({ message, type: 'error' }),
    warning: (message) => toastService.show({ message, type: 'warning' }),
};
const icons = {
    success: 'success', error: 'error', warning: 'warning',
};
export const Toast = {
    view() {
        if (toasts.length === 0)
            return null;
        return m('.toast-container', toasts.map(toast => m('.toast-item', { key: toast.id, class: toast.type }, [
            Icon(icons[toast.type]),
            m('.msg', toast.message),
        ])));
    },
};
