#[cfg(desktop)]
mod tray;
mod server;

use std::sync::{Arc, Mutex};
use tauri::{Emitter, Manager};
use tauri_plugin_shell::ShellExt;

pub fn run() {
    let server_process = Arc::new(Mutex::new(None::<tauri_plugin_shell::process::CommandChild>));
    let server_process_for_setup = Arc::clone(&server_process);

    tauri::Builder::default()
        .plugin(tauri_plugin_updater::Builder::new().build())
        .plugin(tauri_plugin_shell::init())
        .setup(move |app| {
            #[cfg(all(desktop))]
            {
                let handle = app.handle();
                tray::create_tray(handle)?;
            }

            let window = app.get_webview_window("main").unwrap();

            // Open dev tools only when in dev mode
            #[cfg(debug_assertions)]
            {
                window.open_devtools();
            }

            server::launch_seanime_server(app.handle().clone(), server_process_for_setup);
            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while running tauri application")
        .run({
            let server_process_for_exit = Arc::clone(&server_process);
            move |app, event| {
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

