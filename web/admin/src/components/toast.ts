import m from 'mithril';
import { Icon, IconName } from '../icons';

interface ToastConfig {
  message: string;
  type: 'success' | 'error' | 'warning';
  duration?: number;
}

let toasts: (ToastConfig & { id: number })[] = [];
let nextId = 0;

export const toastService = {
  show(config: ToastConfig) {
    const id = nextId++;
    toasts.push({ ...config, id });
    m.redraw();
    setTimeout(() => {
      toasts = toasts.filter(t => t.id !== id);
      m.redraw();
    }, config.duration || 3000);
  },

  success: (message: string) => toastService.show({ message, type: 'success' }),
  error: (message: string) => toastService.show({ message, type: 'error' }),
  warning: (message: string) => toastService.show({ message, type: 'warning' }),
};

const icons: Record<string, IconName> = {
  success: 'success', error: 'error', warning: 'warning',
};

export const Toast: m.Component = {
  view() {
    if (toasts.length === 0) return null;

    return m('.toast-container', toasts.map(toast =>
      m('.toast-item', { key: toast.id, class: toast.type }, [
        Icon(icons[toast.type]),
        m('.msg', toast.message),
      ]),
    ));
  },
};
