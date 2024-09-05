use std::sync::{Arc, Mutex};
use strip_ansi_escapes;
use tauri::{AppHandle, Emitter, Listener, Manager};
use tauri_plugin_shell::process::CommandEvent;
use tauri_plugin_shell::ShellExt;
use tokio::time::{sleep, Duration};
use crate::constants::{CRASH_SCREEN_WINDOW_LABEL, MAIN_WINDOW_LABEL, SPLASHSCREEN_WINDOW_LABEL};

pub fn launch_seanime_server(
    app: AppHandle,
    child_process: Arc<Mutex<Option<tauri_plugin_shell::process::CommandChild>>>,
) {
    tauri::async_runtime::spawn(async move {
        let main_window = app.get_webview_window(MAIN_WINDOW_LABEL).unwrap();
        let splashscreen = app.get_webview_window(SPLASHSCREEN_WINDOW_LABEL).unwrap();
        let crash_screen = app.get_webview_window(CRASH_SCREEN_WINDOW_LABEL).unwrap();

        println!("Starting Seanime, {}", env!("TEST_DATADIR"));

        let mut sidecar_command = app.shell().sidecar("seanime").unwrap();

        // Use test data dir during development
        #[cfg(debug_assertions)]
        {
            sidecar_command = sidecar_command.args(["-datadir", env!("TEST_DATADIR")]);
        }

        let (mut rx, child) = match sidecar_command.spawn() {
            Ok(result) => result,
            Err(e) => {
                // Seanime server failed to open -> close splashscreen and display crash screen
                splashscreen.close().unwrap();
                crash_screen.show().unwrap();
                // Listen for the "crash-screen-loaded" event before emitting the crash message
                let crash_screen_ = crash_screen.clone();
                crash_screen.once("crash-screen-loaded", move |_| {
                    crash_screen_.emit("crash", format!("The server failed to start: {}. Closing in 10 seconds.", e)).expect("failed to emit event");
                });
                sleep(Duration::from_secs(10)).await;
                std::process::exit(1);
            }
        };

        // Store the child process
        *child_process.lock().unwrap() = Some(child);

        // Read server terminal output
        while let Some(event) = rx.recv().await {
            match event {
                CommandEvent::Stdout(line) => {
                    let line_without_colors = strip_ansi_escapes::strip(line);
                    match String::from_utf8(line_without_colors) {
                        Ok(line_str) => {
                            if line_str.contains("Seanime started at") {
                                sleep(Duration::from_secs(2)).await;

                                splashscreen.close().unwrap();
                                main_window.maximize().unwrap();
                                main_window.show().unwrap();
                            }
                            // Emit the line to the main window
                            main_window
                                .emit("message", Some(format!("{}", line_str)))
                                .expect("failed to emit event");
                        }
                        Err(_) => {}
                    }
                }
                CommandEvent::Terminated(status) => {
                    // If the server process terminates, exit the Tauri app
                    eprintln!(
                        "Seanime server process terminated with status: {:?}",
                        status
                    );
                    splashscreen.close().unwrap();
                    #[cfg(debug_assertions)]
                    {
                        main_window.close_devtools();
                    }
                    main_window.close().unwrap();
                    crash_screen.show().unwrap();

                    #[cfg(debug_assertions)]
                    {
                        crash_screen.open_devtools();
                    }


                    app.emit("crash", format!("Seanime server process terminated with status: {}. Closing in 10 seconds.", status.code.unwrap_or(1))).expect("failed to emit event");

                    sleep(Duration::from_secs(10)).await;
                    app.exit(1);
                    break;

                }
                _ => {}
            }
        }
    });
}
