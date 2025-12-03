// Web Worker for PGS rendering
let ctx = null
let offscreenCanvas = null
const events = new Map()
const imageCache = new Map()
let currentEvent = null
let currentEventRendered = false
let timeOffset = 0
let debug = false

function getEventKey(event) {
    return `${event.startTime}-${event.duration}-${event.imageData.substring(0, 50)}`
}

async function preloadImage(base64Data) {
    const response = await fetch(base64Data)
    const blob = await response.blob()
    return createImageBitmap(blob)
}

function logDebug(message, data) {
    if (debug) {
        self.postMessage({
            type: "debug",
            payload: { message, data },
        })
    }
}

async function handleAddEvent(event) {
    const key = getEventKey(event)

    if (events.has(key)) {
        return
    }

    events.set(key, event)

    try {
        const img = await preloadImage(event.imageData)
        imageCache.set(event.imageData, img)
        logDebug("Preloaded image", {
            startTime: event.startTime,
            width: img.width,
            height: img.height,
        })
    } catch (err) {
        self.postMessage({
            type: "error",
            payload: { message: "Failed to preload image", error: err },
        })
    }

    logDebug("Added PGS event", {
        startTime: event.startTime,
        endTime: event.startTime + event.duration,
        duration: event.duration,
        width: event.width,
        height: event.height,
        x: event.x,
        y: event.y,
        canvasWidth: event.canvasWidth,
        canvasHeight: event.canvasHeight,
    })
}

function handleRender(payload) {
    if (!ctx || !offscreenCanvas) {
        return
    }

    const currentTime = payload.currentTime + timeOffset

    // Find the event that should be displayed at current time
    let eventToDisplay = null

    for (const event of events.values()) {
        const startTime = event.startTime
        const endTime = event.startTime + event.duration

        if (currentTime >= startTime && currentTime <= endTime) {
            eventToDisplay = event
            break
        }
    }

    // If event changed, clear canvas
    if (eventToDisplay !== currentEvent) {
        ctx.clearRect(0, 0, offscreenCanvas.width, offscreenCanvas.height)
        currentEventRendered = false

        if (eventToDisplay) {
            logDebug("Displaying new PGS event", {
                currentTime,
                startTime: eventToDisplay.startTime,
                endTime: eventToDisplay.startTime + eventToDisplay.duration,
                canvasWidth: offscreenCanvas.width,
                canvasHeight: offscreenCanvas.height,
            })
        } else if (currentEvent) {
            logDebug("Cleared PGS event", { currentTime })
        }

        currentEvent = eventToDisplay
    }

    // Render current event only if it hasn't been rendered yet
    if (eventToDisplay && !currentEventRendered) {
        renderEvent(eventToDisplay, payload.canvasWidth, payload.canvasHeight)
        currentEventRendered = true
    } else if (!eventToDisplay) {
        ctx.clearRect(0, 0, offscreenCanvas.width, offscreenCanvas.height)
        currentEventRendered = false
    }
}

function renderEvent(event, canvasWidth, canvasHeight) {
    if (!ctx || !offscreenCanvas) {
        return
    }

    const img = imageCache.get(event.imageData)
    if (!img) {
        return
    }

    // Video canvas dimensions from event or use canvas size
    const videoCanvasWidth = event.canvasWidth || canvasWidth
    const videoCanvasHeight = event.canvasHeight || canvasHeight

    // Scale factors
    const scaleX = canvasWidth / videoCanvasWidth
    const scaleY = canvasHeight / videoCanvasHeight

    // Position (default to bottom center if not specified)
    let x = event.x !== undefined ? event.x : (videoCanvasWidth - event.width) / 2
    let y = event.y !== undefined ? event.y : videoCanvasHeight - event.height - 20

    // Apply scaling
    x *= scaleX
    y *= scaleY

    const width = event.width * scaleX
    const height = event.height * scaleY

    logDebug("Rendering PGS image", {
        x,
        y,
        width,
        height,
        scaleX,
        scaleY,
        imgWidth: img.width,
        imgHeight: img.height,
    })

    // Handle cropping if specified
    if (event.cropX !== undefined && event.cropY !== undefined &&
        event.cropWidth !== undefined && event.cropHeight !== undefined) {

        const sx = event.cropX
        const sy = event.cropY
        const sWidth = event.cropWidth
        const sHeight = event.cropHeight

        ctx.drawImage(
            img,
            sx, sy, sWidth, sHeight,
            x, y, width, height,
        )
    } else {
        ctx.drawImage(img, x, y, width, height)
    }

    if (debug) {
        // Draw translucent overlay over entire canvas when subtitle is present
        ctx.fillStyle = "rgba(255, 0, 255, 0.1)"
        ctx.fillRect(0, 0, canvasWidth, canvasHeight)

        // Draw debug border around subtitle
        ctx.strokeStyle = "purple"
        ctx.lineWidth = 2
        ctx.strokeRect(x, y, width, height)
    }
}

function handleResize(payload) {
    if (!offscreenCanvas) {
        return
    }

    offscreenCanvas.width = payload.width
    offscreenCanvas.height = payload.height
    currentEventRendered = false

    logDebug("Resized canvas", {
        width: payload.width,
        height: payload.height,
    })
}

function handleClear() {
    events.clear()
    imageCache.clear()
    currentEvent = null
    currentEventRendered = false

    if (ctx && offscreenCanvas) {
        ctx.clearRect(0, 0, offscreenCanvas.width, offscreenCanvas.height)
    }
}

self.onmessage = async (e) => {
    const { type, payload } = e.data

    switch (type) {
        case "init":
            offscreenCanvas = payload.canvas
            if (offscreenCanvas) {
                ctx = offscreenCanvas.getContext("2d", { alpha: true })
                debug = payload.debug ?? false
                logDebug("Worker initialized", {
                    width: offscreenCanvas.width,
                    height: offscreenCanvas.height,
                })
            }
            break

        case "addEvent":
            await handleAddEvent(payload)
            break

        case "render":
            handleRender(payload)
            break

        case "resize":
            handleResize(payload)
            break

        case "clear":
            handleClear()
            break

        case "setTimeOffset":
            timeOffset = payload.offset
            break

        case "setDebug":
            debug = payload.debug
            break
    }
}

