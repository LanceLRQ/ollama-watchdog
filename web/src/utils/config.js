import { DefaultServerSettings } from '@/models/server'

export const loadSettingsFromLocalStorage = () => {
    // 从本地存储加载设置
    const settings = localStorage.getItem("watchdog_api_settings");
    if (settings) {
        try {
            return Object.assign(
                {},
                DefaultServerSettings,
                JSON.parse(settings)
            );
        } catch (e) {
            console.error("Error parsing settings:", e);
        }
    }
    return { ... DefaultServerSettings };
};

export const saveSettingsToLocalStorage = (settingsForm) => {
    // 将设置保存到本地存储
    localStorage.setItem("watchdog_api_settings", JSON.stringify(settingsForm));
}