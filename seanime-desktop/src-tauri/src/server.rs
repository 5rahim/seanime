use std::sync::{Arc, Mutex};
use strip_ansi_escapes;
use tauri::{AppHandle, Emitter, Manager};
use tauri_plugin_shell::process::CommandEvent;
use tauri_plugin_shell::ShellExt;

pub fn launch_seanime_server(app: AppHandle, child_process: Arc<Mutex<Option<tauri_plugin_shell::process::CommandChild>>>) {
    tauri::async_runtime::spawn(async move {
        let main_window = app.get_webview_window("main").unwrap();

        println!("Starting Seanime, {}", env!("TEST_DATADIR"));

        let mut sidecar_command = app.shell()
            .sidecar("seanime")
            .unwrap();

        // Use test data dir during development
        #[cfg(debug_assertions)]
        {
            sidecar_command = sidecar_command.args(["-datadir", env!("TEST_DATADIR")]);
        }

        let (mut rx, child) = match sidecar_command.spawn() {
            Ok(result) => result,
            Err(_) => {
                // Close the app if server launch fails
                std::process::exit(1);
            }
        };

        // Store the child process
        *child_process.lock().unwrap() = Some(child);

        let mut seanime_started = false;

        // Read server terminal output
        while let Some(event) = rx.recv().await {
            match event {
                CommandEvent::Stdout(line) => {
                    let line_without_colors = strip_ansi_escapes::strip(line);
                    match String::from_utf8(line_without_colors) {
                        Ok(line_str) => {
                            if line_str.contains("Seanime started at") {
                                seanime_started = true;
                            }
                            // Emit the line to the main window
                            main_window.emit("message", Some(format!("{}", line_str)))
                                .expect("failed to emit event");
                        }
                        Err(_) => {}
                    }
                }
                CommandEvent::Terminated(status) => {
                    // If the server process terminates, exit the Tauri app
                    eprintln!("Seanime server process terminated with status: {:?}", status);
                    app.exit(1);
                    break;
                }
                _ => {}
            }
        }
    });
}
