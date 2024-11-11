use crate::constants::MAIN_WINDOW_LABEL;
use tauri::{
    menu::{Menu, MenuItem},
    tray::{MouseButton, MouseButtonState, TrayIconBuilder, TrayIconEvent},
    Manager, Runtime,
};

pub fn create_tray<R: Runtime>(app: &tauri::AppHandle<R>) -> tauri::Result<()> {
    let quit_i = MenuItem::with_id(app, "quit", "Quit Seanime", true, None::<&str>)?;
    let toggle_visibility_i = MenuItem::with_id(app, "toggle_visibility", "Toggle visibility", true, None::<&str>)?;


    let menu = Menu::with_items(app, &[&toggle_visibility_i, &quit_i])?;

    let _ = TrayIconBuilder::with_id("tray")
        .icon(app.default_window_icon().unwrap().clone())
        .menu(&menu)
        .menu_on_left_click(false)
        .on_menu_event(move |app, event| match event.id.as_ref() {
            "quit" => {
                app.exit(0);
            }
            "toggle_visibility" => {
                if let Some(window) = app.get_webview_window(MAIN_WINDOW_LABEL) {
                    if window.is_minimized().unwrap() {
                        let _ = window.show();
                        let _ = window.set_focus();
                    } else {
                        let _ = window.minimize();
                    }
                }
            }
            // Add more events here
            _ => {}
        })
        .on_tray_icon_event(|tray, event| {
            if let TrayIconEvent::Click {
                button: MouseButton::Left,
                button_state: MouseButtonState::Up,
                ..
            } = event
            {
                let app = tray.app_handle();
                if let Some(window) = app.get_webview_window(MAIN_WINDOW_LABEL) {
                    let _ = window.show();
                    let _ = window.set_focus();
                }
            }
        })
        .build(app);

    Ok(())
}
