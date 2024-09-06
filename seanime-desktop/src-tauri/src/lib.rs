#[cfg(desktop)]
mod tray;
mod constants;
mod server;

use std::sync::{Arc, Mutex};
#[cfg(target_os = "macos")]
use tauri::utils::TitleBarStyle;
use tauri::{Listener, Manager};
use tauri_plugin_os;
use constants::{MAIN_WINDOW_LABEL};

pub fn run() {
    let server_process = Arc::new(Mutex::new(
        None::<tauri_plugin_shell::process::CommandChild>,
    ));
    let server_process_for_setup = Arc::clone(&server_process);

    tauri::Builder::default()
        .plugin(tauri_plugin_updater::Builder::new().build())
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_os::init())
        .setup(move |app| {
            #[cfg(all(desktop))]
            {
                let handle = app.handle();
                tray::create_tray(handle)?;
            }

            let main_window = app.get_webview_window(MAIN_WINDOW_LABEL).unwrap();
            main_window.hide().unwrap();

            // Set overlay title bar only when building for macOS
            #[cfg(target_os = "macos")]
            main_window.set_title_bar_style(TitleBarStyle::Overlay).unwrap();

            // Hide the title bar on Windows
            #[cfg(any(target_os = "windows"))]
            main_window.set_decorations(false).unwrap();

            // Open dev tools only when in dev mode
            #[cfg(debug_assertions)]
            {
                main_window.open_devtools();
            }

            server::launch_seanime_server(app.handle().clone(), server_process_for_setup);
            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while running tauri application")
        .run({
            let server_process_for_exit = Arc::clone(&server_process);
            move |app, event| {
                let server_process_for_exit_ = Arc::clone(&server_process);
                app.listen("kill-server", move |e| {
                    let mut child_guard = server_process_for_exit_.lock().unwrap();
                    if let Some(child) = child_guard.take() {
                        // Kill server process
                        if let Err(e) = child.kill() {
                            eprintln!("Failed to kill server process: {}", e);
                        }
                    }
                });

                match event {
                    tauri::RunEvent::WindowEvent {
                        label,
                        event: tauri::WindowEvent::CloseRequested { api, .. },
                        ..
                    } => {
                        // Hide the window when user clicks 'X'
                        let win = app.get_webview_window(label.as_str()).unwrap();
                        win.hide().unwrap();
                        // Prevent the window from being closed
                        api.prevent_close();
                    }

                    // tauri::RunEvent::Exit => {
                    //     let mut child_guard = server_process_for_exit.lock().unwrap();
                    //     if let Some(child) = child_guard.take() {
                    //         // Kill server process
                    //         if let Err(e) = child.kill() {
                    //             eprintln!("Failed to kill server process: {}", e);
                    //         }
                    //     }
                    // }

                    // The app is about to exit
                    tauri::RunEvent::ExitRequested { .. } => {
                        let mut child_guard = server_process_for_exit.lock().unwrap();
                        if let Some(child) = child_guard.take() {
                            // Kill server process
                            if let Err(e) = child.kill() {
                                eprintln!("Failed to kill server process: {}", e);
                            }
                        }
                    }
                    _ => {}
                }
            }
        });
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
