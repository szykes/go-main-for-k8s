# Go Kubernetes Template

This is a ready-to-use template for Go applications running in Kubernetes. It safely handles shutdowns, catches panics, reads environment variables, builds secure binaries, and logs data in a clean way.

## Features

- **Safe Shutdown**: Correctly handles the Kubernetes pod stop signal (`SIGTERM`). It finishes current tasks before stopping the app.
- **Panic Catching**: Catches all Go panics, ang logs the error safely in JSON format.
- **Secure Build Process**: Uses a `Makefile` to inject build data, remove local file paths, and build an app that can run on an empty `scratch` Docker image.
- **Local vs. Cloud Environments**: Easily loads a `.env` file when you test locally. In Kubernetes, it uses standard ConfigMaps or Secrets.
- **Jobs or Servers**: Works perfectly whether your app is a short task (like a Database Migration Job) or runs infinitely (like a Web Server).

## How It Works

### 1. Panic Handling & Security
If a Go app panics and you do not catch it, it prints a large stack trace to the standard error output. This is bad for production because if you use JSON logs (`slog.NewJSONHandler`), a raw text stack trace will break your log tools.

We fix this by using `recover()` at the top of `main()`. If a panic happens, we catch it, log it as clean JSON, and exit with error code `2`.

### 2. Safe Signal Handling
When Kubernetes wants to stop a pod, it sends a `SIGTERM` signal. If your app does not stop in time, Kubernetes kills it forcefully with `SIGKILL`.

This app listens for `SIGTERM` and `SIGINT`. When the signal comes, it calls `stop()` on the `context`. This tells the app to finish its work and shut down cleanly.

> [!IMPORTANT]
> When your app gets a `SIGTERM` and stops cleanly, it exits with code non-zero, your monitoring tools will think the app crashed. This causes false alarms every time you update your app in Kubernetes! However, when something happens during production, this is the best way to notify monitoring tools.

### 3. Secure Building & Build Information
To help engineers fix bugs, `buildinfo.go` logs the app version, commit hash, and build time when the app starts. We use a `Makefile` to build the app safely and correctly:

* **Injecting Build Info**: The `Makefile` gets the current Git commit and the current date. It uses `-ldflags="-X ..."` to push these values directly into the Go variables when the app is built. You never have to type them by hand.
* **`CGO_ENABLED=0`**: This tells Go not to use C libraries. This means the final app file is completely standalone. Because of this, you can put the app inside a Docker `scratch` image (an image with literally nothing else in it). This makes your Docker image very small and very secure.
* **`-trimpath`**: Normally, Go saves the names of your computer's folders inside the compiled app. We use `-trimpath` to delete your local paths from the final file. This is an extra security step so you do not leak your local computer structure.

### 4. Support for Jobs and Daemons
This code handles two main ways to run apps in Kubernetes:
- **Jobs / InitContainers**: The `app.Run()` function finishes its work and returns `nil`. The app then exits nicely with code `0` (Success).
- **Servers / Daemons**: The `app.Run()` function runs forever. It only stops when Kubernetes sends a system signal to cancel the `context`.

## Usage

### Using the Makefile

The included `Makefile` makes it easy to run and build the app safely.

**Run locally (for testing):**
This will automatically load the `.env` file, turn on `DEBUG_MODE`, and run the app.
```bash
make run-app-debug
```
