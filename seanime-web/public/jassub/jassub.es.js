typeof HTMLVideoElement < "u" && !("requestVideoFrameCallback" in HTMLVideoElement.prototype) && "getVideoPlaybackQuality" in HTMLVideoElement.prototype && (HTMLVideoElement.prototype._rvfcpolyfillmap = {}, HTMLVideoElement.prototype.requestVideoFrameCallback = function (c) {
    const e = performance.now(), t = this.getVideoPlaybackQuality(),
        s = this.mozPresentedFrames || this.mozPaintedFrames || t.totalVideoFrames - t.droppedVideoFrames, a = (n, r) => {
            const i = this.getVideoPlaybackQuality(), o = this.mozPresentedFrames || this.mozPaintedFrames || i.totalVideoFrames - i.droppedVideoFrames;
            if (o > s) {
                const d = this.mozFrameDelay || i.totalFrameDelay - t.totalFrameDelay || 0, l = r - n;
                c(r, {
                    presentationTime: r + d * 1e3,
                    expectedDisplayTime: r + l,
                    width: this.videoWidth,
                    height: this.videoHeight,
                    mediaTime: Math.max(0, this.currentTime || 0) + l / 1e3,
                    presentedFrames: o,
                    processingDuration: d
                }), delete this._rvfcpolyfillmap[e];
            } else
                this._rvfcpolyfillmap[e] = requestAnimationFrame((d) => a(r, d));
        };
    return this._rvfcpolyfillmap[e] = requestAnimationFrame((n) => a(e, n)), e;
}, HTMLVideoElement.prototype.cancelVideoFrameCallback = function (c) {
    cancelAnimationFrame(this._rvfcpolyfillmap[c]), delete this._rvfcpolyfillmap[c];
});
const _ = {
    bt709: "BT709",
    // these might not be exactly correct? oops?
    bt470bg: "BT601",
    // alias BT.601 PAL... whats the difference?
    smpte170m: "BT601"
    // alias BT.601 NTSC... whats the difference?
}, f = {
    BT601: {
        BT709: "1.0863 -0.0723 -0.014 0 0 0.0965 0.8451 0.0584 0 0 -0.0141 -0.0277 1.0418"
    },
    BT709: {
        BT601: "0.9137 0.0784 0.0079 0 0 -0.1049 1.1722 -0.0671 0 0 0.0096 0.0322 0.9582"
    },
    FCC: {
        BT709: "1.0873 -0.0736 -0.0137 0 0 0.0974 0.8494 0.0531 0 0 -0.0127 -0.0251 1.0378",
        BT601: "1.001 -0.0008 -0.0002 0 0 0.0009 1.005 -0.006 0 0 0.0013 0.0027 0.996"
    },
    SMPTE240M: {
        BT709: "0.9993 0.0006 0.0001 0 0 -0.0004 0.9812 0.0192 0 0 -0.0034 -0.0114 1.0148",
        BT601: "0.913 0.0774 0.0096 0 0 -0.1051 1.1508 -0.0456 0 0 0.0063 0.0207 0.973"
    }
};

class h extends EventTarget {
    // test support for WASM, ImageData, alphaBug, but only once, on init so it doesn't run when first running the page
    /** @type {boolean|null} */
    static _supportsSIMD = null;
    /** @type {boolean|null} */
    static _hasAlphaBug = null;
    /** @type {boolean|null} */
    static _hasBitmapBug = null;

    /**
     * @param {Object} options Settings object.
     * @param {HTMLVideoElement} options.video Video to use as target for rendering and event listeners. Optional if canvas is specified instead.
     * @param {HTMLCanvasElement} [options.canvas=HTMLCanvasElement] Canvas to use for manual handling. Not required if video is specified.
     * @param {'js'|'wasm'} [options.blendMode='js'] Which image blending mode to use. WASM will perform better on lower end devices, JS will perform better if the device and browser supports hardware acceleration.
     * @param {Boolean} [options.asyncRender=true] Whether or not to use async rendering, which offloads the CPU by creating image bitmaps on the GPU.
     * @param {Boolean} [options.offscreenRender=true] Whether or not to render things fully on the worker, greatly reduces CPU usage.
     * @param {Boolean} [options.onDemandRender=true] Whether or not to render subtitles as the video player decodes renders, rather than predicting which frame the player is on using events.
     * @param {Number} [options.targetFps=24] Target FPS to render subtitles at. Ignored when onDemandRender is enabled.
     * @param {Number} [options.timeOffset=0] Subtitle time offset in seconds.
     * @param {Boolean} [options.debug=false] Whether or not to print debug information.
     * @param {Number} [options.prescaleFactor=1.0] Scale down (< 1.0) the subtitles canvas to improve performance at the expense of quality, or scale it up (> 1.0).
     * @param {Number} [options.prescaleHeightLimit=1080] The height in pixels beyond which the subtitles canvas won't be prescaled.
     * @param {Number} [options.maxRenderHeight=0] The maximum rendering height in pixels of the subtitles canvas. Beyond this subtitles will be upscaled by the browser.
     * @param {Boolean} [options.dropAllAnimations=false] Attempt to discard all animated tags. Enabling this may severly mangle complex subtitles and should only be considered as an last ditch effort of uncertain success for hardware otherwise incapable of displaing anything. Will not reliably work with manually edited or allocated events.
     * @param {Boolean} [options.dropAllBlur=false] The holy grail of performance gains. If heavy TS lags a lot, disabling this will make it ~x10 faster. This drops blur from all added subtitle tracks making most text and backgrounds look sharper, this is way less intrusive than dropping all animations, while still offering major performance gains.
     * @param {String} [options.workerUrl='jassub-worker.js'] The URL of the worker.
     * @param {String} [options.wasmUrl='jassub-worker.wasm'] The URL of the worker WASM.
     * @param {String} [options.legacyWasmUrl='jassub-worker.wasm.js'] The URL of the worker WASM. Only loaded if the browser doesn't support WASM.
     * @param {String} options.modernWasmUrl The URL of the modern worker WASM. This includes faster ASM instructions, but is only supported by newer browsers, disabled if the URL isn't defined.
     * @param {String} [options.subUrl=options.subContent] The URL of the subtitle file to play.
     * @param {String} [options.subContent=options.subUrl] The content of the subtitle file to play.
     * @param {String[]|Uint8Array[]} [options.fonts] An array of links or Uint8Arrays to the fonts used in the subtitle. If Uint8Array is used the array is copied, not referenced. This forces all the fonts in this array to be loaded by the renderer, regardless of if they are used.
     * @param {Object} [options.availableFonts={'liberation sans': './default.woff2'}] Object with all available fonts - Key is font family in lower case, value is link or Uint8Array: { arial: '/font1.ttf' }. These fonts are selectively loaded if detected as used in the current subtitle track.
     * @param {String} [options.fallbackFont='liberation sans'] The font family key of the fallback font in availableFonts to use if the other font for the style is missing special glyphs or unicode.
     * @param {Boolean} [options.useLocalFonts=false] If the Local Font Access API is enabled [chrome://flags/#font-access], the library will query for permissions to use local fonts and use them if any are missing. The permission can be queried beforehand using navigator.permissions.request({ name: 'local-fonts' }).
     * @param {Number} [options.libassMemoryLimit] libass bitmap cache memory limit in MiB (approximate).
     * @param {Number} [options.libassGlyphLimit] libass glyph cache memory limit in MiB (approximate).
     */
    constructor(e) {
        if (super(), !globalThis.Worker)
            throw this.destroy("Worker not supported");
        if (!e)
            throw this.destroy("No options provided");
        this._loaded = /** @type {Promise<void>} */
            new Promise((s) => {
                this._init = s;
            });
        const t = h._test();
        if (this._onDemandRender = "requestVideoFrameCallback" in HTMLVideoElement.prototype && (e.onDemandRender ?? !0), this._offscreenRender = "transferControlToOffscreen" in HTMLCanvasElement.prototype && !e.canvas && (e.offscreenRender ?? !0), this.timeOffset = e.timeOffset || 0, this._video = e.video, this._videoHeight = 0, this._videoWidth = 0, this._videoColorSpace = null, this._canvas = e.canvas, this._video && !this._canvas)
            this._canvasParent = document.createElement("div"), this._canvasParent.className = "JASSUB", this._canvasParent.style.position = "relative", this._canvas = this._createCanvas(), this._video.insertAdjacentElement("afterend", this._canvasParent);
        else if (!this._canvas)
            throw this.destroy("Don't know where to render: you should give video or canvas in options.");
        if (this._bufferCanvas = document.createElement("canvas"), this._bufferCtx = this._bufferCanvas.getContext("2d"), !this._bufferCtx)
            throw this.destroy("Canvas rendering not supported");
        this._canvasctrl = this._offscreenRender ? this._canvas.transferControlToOffscreen() : this._canvas, this._ctx = !this._offscreenRender && this._canvasctrl.getContext("2d"), this._lastRenderTime = 0, this.debug = !!e.debug, this.prescaleFactor = e.prescaleFactor || 1, this.prescaleHeightLimit = e.prescaleHeightLimit || 1080, this.maxRenderHeight = e.maxRenderHeight || 0, this._boundResize = this.resize.bind(this), this._boundTimeUpdate = this._timeupdate.bind(this), this._boundSetRate = this.setRate.bind(this), this._boundUpdateColorSpace = this._updateColorSpace.bind(this), this._video && this.setVideo(e.video), this._onDemandRender && (this.busy = !1, this._lastDemandTime = null), this._worker = new Worker(e.workerUrl || "jassub-worker.js"), this._worker.onmessage = (s) => this._onmessage(s), this._worker.onerror = (s) => this._error(s), t.then(() => {
            this._worker.postMessage({
                target: "init",
                wasmUrl: h._supportsSIMD && e.modernWasmUrl ? e.modernWasmUrl : e.wasmUrl || "jassub-worker.wasm",
                legacyWasmUrl: e.legacyWasmUrl || "jassub-worker.wasm.js",
                asyncRender: typeof createImageBitmap < "u" && (e.asyncRender ?? !0),
                onDemandRender: this._onDemandRender,
                width: this._canvasctrl.width || 0,
                height: this._canvasctrl.height || 0,
                blendMode: e.blendMode || "js",
                subUrl: e.subUrl,
                subContent: e.subContent || null,
                fonts: e.fonts || [],
                availableFonts: e.availableFonts || {"liberation sans": "./default.woff2"},
                fallbackFont: e.fallbackFont || "liberation sans",
                debug: this.debug,
                targetFps: e.targetFps || 24,
                dropAllAnimations: e.dropAllAnimations,
                dropAllBlur: e.dropAllBlur,
                libassMemoryLimit: e.libassMemoryLimit || 0,
                libassGlyphLimit: e.libassGlyphLimit || 0,
                // @ts-ignore
                useLocalFonts: typeof queryLocalFonts < "u" && (e.useLocalFonts ?? !0),
                hasBitmapBug: h._hasBitmapBug
            }), this._offscreenRender === !0 && this.sendMessage("offscreenCanvas", null, [this._canvasctrl]);
        });
    }

    static async _test() {
        if (h._hasBitmapBug !== null)
            return null;
        try {
            h._supportsSIMD = WebAssembly.validate(Uint8Array.of(0, 97, 115, 109, 1, 0, 0, 0, 1, 5, 1, 96, 0, 1, 123, 3, 2, 1, 0, 10, 10, 1, 8, 0, 65, 0, 253, 15, 253, 98, 11));
        } catch {
            h._supportsSIMD = !1;
        }
        const e = document.createElement("canvas"), t = e.getContext("2d", {willReadFrequently: !0});
        if (!t)
            throw new Error("Canvas rendering not supported");
        if (typeof ImageData.prototype.constructor == "function")
            try {
                new ImageData(new Uint8ClampedArray([0, 0, 0, 0]), 1, 1);
            } catch {
                console.log("Detected that ImageData is not constructable despite browser saying so"), self.ImageData = function (o, d, l) {
                    const m = t.createImageData(d, l);
                    return o && m.data.set(o), m;
                };
            }
        const s = document.createElement("canvas"), a = s.getContext("2d", {willReadFrequently: !0});
        if (!a)
            throw new Error("Canvas rendering not supported");
        e.width = s.width = 1, e.height = s.height = 1, t.clearRect(0, 0, 1, 1), a.clearRect(0, 0, 1, 1);
        const n = a.getImageData(0, 0, 1, 1).data;
        t.putImageData(new ImageData(new Uint8ClampedArray([0, 255, 0, 0]), 1, 1), 0, 0), a.drawImage(e, 0, 0);
        const r = a.getImageData(0, 0, 1, 1).data;
        if (h._hasAlphaBug = n[1] !== r[1], h._hasAlphaBug && console.log("Detected a browser having issue with transparent pixels, applying workaround"), typeof createImageBitmap < "u") {
            const i = new Uint8ClampedArray([255, 0, 255, 0, 255]).subarray(1, 5);
            a.drawImage(await createImageBitmap(new ImageData(i, 1)), 0, 0);
            const {data: o} = a.getImageData(0, 0, 1, 1);
            h._hasBitmapBug = !1;
            for (const [d, l] of o.entries())
                if (Math.abs(i[d] - l) > 15) {
                    h._hasBitmapBug = !0, console.log("Detected a browser having issue with partial bitmaps, applying workaround");
                    break;
                }
        } else
            h._hasBitmapBug = !1;
        e.remove(), s.remove();
    }

    _createCanvas() {
        return this._canvas = document.createElement("canvas"), this._canvas.style.display = "block", this._canvas.style.position = "absolute", this._canvas.style.pointerEvents = "none", this._canvasParent.appendChild(this._canvas), this._canvas;
    }

    /**
     * Resize the canvas to given parameters. Auto-generated if values are ommited.
     * @param  {Number} [width=0]
     * @param  {Number} [height=0]
     * @param  {Number} [top=0]
     * @param  {Number} [left=0]
     * @param  {Boolean} [force=false]
     */
    resize(e = 0, t = 0, s = 0, a = 0, n = this._video?.paused) {
        if ((!e || !t) && this._video) {
            const r = this._getVideoPosition();
            let i = null;
            if (this._videoWidth) {
                const o = this._video.videoWidth / this._videoWidth, d = this._video.videoHeight / this._videoHeight;
                i = this._computeCanvasSize((r.width || 0) / o, (r.height || 0) / d);
            } else
                i = this._computeCanvasSize(r.width || 0, r.height || 0);
            e = i.width, t = i.height, this._canvasParent && (s = r.y - (this._canvasParent.getBoundingClientRect().top - this._video.getBoundingClientRect().top), a = r.x), this._canvas.style.width = r.width + "px", this._canvas.style.height = r.height + "px";
        }
        this._canvas.style.top = s + "px", this._canvas.style.left = a + "px", n && this.busy === !1 ? this.busy = !0 : n = !1, this.sendMessage("canvas", {
            width: e,
            height: t,
            force: n
        });
    }

    _getVideoPosition(e = this._video.videoWidth, t = this._video.videoHeight) {
        const s = e / t, {offsetWidth: a, offsetHeight: n} = this._video, r = a / n;
        e = a, t = n, r > s ? e = Math.floor(n * s) : t = Math.floor(a / s);
        const i = (a - e) / 2, o = (n - t) / 2;
        return {width: e, height: t, x: i, y: o};
    }

    _computeCanvasSize(e = 0, t = 0) {
        const s = this.prescaleFactor <= 0 ? 1 : this.prescaleFactor, a = self.devicePixelRatio || 1;
        if (e = e * a, t = t * a, t <= 0 || e <= 0)
            e = 0, t = 0;
        else {
            const n = s < 1 ? -1 : 1;
            let r = t * a;
            n * r * s <= n * this.prescaleHeightLimit ? r *= s : n * r < n * this.prescaleHeightLimit && (r = this.prescaleHeightLimit), this.maxRenderHeight > 0 && r > this.maxRenderHeight && (r = this.maxRenderHeight), e *= r / t, t = r;
        }
        return {width: e, height: t};
    }

    _timeupdate({type: e}) {
        const s = {
            seeking: !0,
            waiting: !0,
            playing: !1
        }[e];
        s != null && (this._playstate = s), this.setCurrentTime(this._video.paused || this._playstate, this._video.currentTime + this.timeOffset);
    }

    /**
     * Change the video to use as target for event listeners.
     * @param  {HTMLVideoElement} video
     */
    setVideo(e) {
        e instanceof HTMLVideoElement ? (this._removeListeners(), this._video = e, this._onDemandRender ? this._video.requestVideoFrameCallback(this._handleRVFC.bind(this)) : (this._playstate = e.paused, e.addEventListener("timeupdate", this._boundTimeUpdate, !1), e.addEventListener("progress", this._boundTimeUpdate, !1), e.addEventListener("waiting", this._boundTimeUpdate, !1), e.addEventListener("seeking", this._boundTimeUpdate, !1), e.addEventListener("playing", this._boundTimeUpdate, !1), e.addEventListener("ratechange", this._boundSetRate, !1), e.addEventListener("resize", this._boundResize, !1)), "VideoFrame" in window && (e.addEventListener("loadedmetadata", this._boundUpdateColorSpace, !1), e.readyState > 2 && this._updateColorSpace()), e.videoWidth > 0 && this.resize(), typeof ResizeObserver < "u" && (this._ro || (this._ro = new ResizeObserver(() => this.resize())), this._ro.observe(e))) : this._error("Video element invalid!");
    }

    runBenchmark() {
        this.sendMessage("runBenchmark");
    }

    /**
     * Overwrites the current subtitle content.
     * @param  {String} url URL to load subtitles from.
     */
    setTrackByUrl(e) {
        this.sendMessage("setTrackByUrl", {url: e}), this._reAttachOffscreen(), this._ctx && (this._ctx.filter = "none");
    }

    /**
     * Overwrites the current subtitle content.
     * @param  {String} content Content of the ASS file.
     */
    setTrack(e) {
        this.sendMessage("setTrack", {content: e}), this._reAttachOffscreen(), this._ctx && (this._ctx.filter = "none");
    }

    /**
     * Free currently used subtitle track.
     */
    freeTrack() {
        this.sendMessage("freeTrack");
    }

    /**
     * Sets the playback state of the media.
     * @param  {Boolean} isPaused Pause/Play subtitle playback.
     */
    setIsPaused(e) {
        this.sendMessage("video", {isPaused: e});
    }

    /**
     * Sets the playback rate of the media [speed multiplier].
     * @param  {Number} rate Playback rate.
     */
    setRate(e) {
        this.sendMessage("video", {rate: e});
    }

    /**
     * Sets the current time, playback state and rate of the subtitles.
     * @param  {Boolean} [isPaused] Pause/Play subtitle playback.
     * @param  {Number} [currentTime] Time in seconds.
     * @param  {Number} [rate] Playback rate.
     */
    setCurrentTime(e, t, s) {
        this.sendMessage("video", {isPaused: e, currentTime: t, rate: s, colorSpace: this._videoColorSpace});
    }

    /**
     * @typedef {Object} ASS_Event
     * @property {Number} Start Start Time of the Event, in 0:00:00:00 format ie. Hrs:Mins:Secs:hundredths. This is the time elapsed during script playback at which the text will appear onscreen. Note that there is a single digit for the hours!
     * @property {Number} Duration End Time of the Event, in 0:00:00:00 format ie. Hrs:Mins:Secs:hundredths. This is the time elapsed during script playback at which the text will disappear offscreen. Note that there is a single digit for the hours!
     * @property {String} Style Style name. If it is "Default", then your own *Default style will be subtituted.
     * @property {String} Name Character name. This is the name of the character who speaks the dialogue. It is for information only, to make the script is easier to follow when editing/timing.
     * @property {Number} MarginL 4-figure Left Margin override. The values are in pixels. All zeroes means the default margins defined by the style are used.
     * @property {Number} MarginR 4-figure Right Margin override. The values are in pixels. All zeroes means the default margins defined by the style are used.
     * @property {Number} MarginV 4-figure Bottom Margin override. The values are in pixels. All zeroes means the default margins defined by the style are used.
     * @property {String} Effect Transition Effect. This is either empty, or contains information for one of the three transition effects implemented in SSA v4.x
     * @property {String} Text Subtitle Text. This is the actual text which will be displayed as a subtitle onscreen. Everything after the 9th comma is treated as the subtitle text, so it can include commas.
     * @property {Number} ReadOrder Number in order of which to read this event.
     * @property {Number} Layer Z-index overlap in which to render this event.
     * @property {Number} _index (Internal) index of the event.
     */
    /**
     * Create a new ASS event directly.
     * @param  {ASS_Event} event
     */
    createEvent(e) {
        this.sendMessage("createEvent", {event: e});
    }

    /**
     * Overwrite the data of the event with the specified index.
     * @param  {ASS_Event} event
     * @param  {Number} index
     */
    setEvent(e, t) {
        this.sendMessage("setEvent", {event: e, index: t});
    }

    /**
     * Remove the event with the specified index.
     * @param  {Number} index
     */
    removeEvent(e) {
        this.sendMessage("removeEvent", {index: e});
    }

    /**
     * Get all ASS events.
     * @param  {function(Error|null, ASS_Event): void} callback Function to callback when worker returns the events.
     */
    getEvents(e) {
        this._fetchFromWorker({
            target: "getEvents"
        }, (t, {events: s}) => {
            e(t, s);
        });
    }

    /**
     * @typedef {Object} ASS_Style
     * @property {String} Name The name of the Style. Case sensitive. Cannot include commas.
     * @property {String} FontName The fontname as used by Windows. Case-sensitive.
     * @property {Number} FontSize Font size.
     * @property {Number} PrimaryColour A long integer BGR (blue-green-red)  value. ie. the byte order in the hexadecimal equivelent of this number is BBGGRR
     * @property {Number} SecondaryColour A long integer BGR (blue-green-red)  value. ie. the byte order in the hexadecimal equivelent of this number is BBGGRR
     * @property {Number} OutlineColour A long integer BGR (blue-green-red)  value. ie. the byte order in the hexadecimal equivelent of this number is BBGGRR
     * @property {Number} BackColour This is the colour of the subtitle outline or shadow, if these are used. A long integer BGR (blue-green-red)  value. ie. the byte order in the hexadecimal equivelent of this number is BBGGRR.
     * @property {Number} Bold This defines whether text is bold (true) or not (false). -1 is True, 0 is False. This is independant of the Italic attribute - you can have have text which is both bold and italic.
     * @property {Number} Italic  Italic. This defines whether text is italic (true) or not (false). -1 is True, 0 is False. This is independant of the bold attribute - you can have have text which is both bold and italic.
     * @property {Number} Underline -1 or 0
     * @property {Number} StrikeOut -1 or 0
     * @property {Number} ScaleX Modifies the width of the font. [percent]
     * @property {Number} ScaleY Modifies the height of the font. [percent]
     * @property {Number} Spacing Extra space between characters. [pixels]
     * @property {Number} Angle The origin of the rotation is defined by the alignment. Can be a floating point number. [degrees]
     * @property {Number} BorderStyle 1=Outline + drop shadow, 3=Opaque box
     * @property {Number} Outline If BorderStyle is 1,  then this specifies the width of the outline around the text, in pixels. Values may be 0, 1, 2, 3 or 4.
     * @property {Number} Shadow If BorderStyle is 1,  then this specifies the depth of the drop shadow behind the text, in pixels. Values may be 0, 1, 2, 3 or 4. Drop shadow is always used in addition to an outline - SSA will force an outline of 1 pixel if no outline width is given.
     * @property {Number} Alignment This sets how text is "justified" within the Left/Right onscreen margins, and also the vertical placing. Values may be 1=Left, 2=Centered, 3=Right. Add 4 to the value for a "Toptitle". Add 8 to the value for a "Midtitle". eg. 5 = left-justified toptitle
     * @property {Number} MarginL This defines the Left Margin in pixels. It is the distance from the left-hand edge of the screen.The three onscreen margins (MarginL, MarginR, MarginV) define areas in which the subtitle text will be displayed.
     * @property {Number} MarginR This defines the Right Margin in pixels. It is the distance from the right-hand edge of the screen. The three onscreen margins (MarginL, MarginR, MarginV) define areas in which the subtitle text will be displayed.
     * @property {Number} MarginV This defines the vertical Left Margin in pixels. For a subtitle, it is the distance from the bottom of the screen. For a toptitle, it is the distance from the top of the screen. For a midtitle, the value is ignored - the text will be vertically centred.
     * @property {Number} Encoding This specifies the font character set or encoding and on multi-lingual Windows installations it provides access to characters used in multiple than one languages. It is usually 0 (zero) for English (Western, ANSI) Windows.
     * @property {Number} treat_fontname_as_pattern
     * @property {Number} Blur
     * @property {Number} Justify
     */
    /**
     * Create a new ASS style directly.
     * @param  {ASS_Style} style
     */
    createStyle(e) {
        this.sendMessage("createStyle", {style: e});
    }

    /**
     * Overwrite the data of the style with the specified index.
     * @param  {ASS_Style} style
     * @param  {Number} index
     */
    setStyle(e, t) {
        this.sendMessage("setStyle", {style: e, index: t});
    }

    /**
     * Remove the style with the specified index.
     * @param  {Number} index
     */
    removeStyle(e) {
        this.sendMessage("removeStyle", {index: e});
    }

    /**
     * Get all ASS styles.
     * @param  {function(Error|null, ASS_Style): void} callback Function to callback when worker returns the styles.
     */
    getStyles(e) {
        this._fetchFromWorker({
            target: "getStyles"
        }, (t, {styles: s}) => {
            e(t, s);
        });
    }

    /**
     * Adds a font to the renderer.
     * @param  {String|Uint8Array} font Font to add.
     */
    addFont(e) {
        this.sendMessage("addFont", {font: e});
    }

    _sendLocalFont(e) {
        try {
            queryLocalFonts().then((t) => {
                const s = t?.find((a) => a.fullName.toLowerCase() === e);
                s && s.blob().then((a) => {
                    a.arrayBuffer().then((n) => {
                        this.addFont(new Uint8Array(n));
                    });
                });
            });
        } catch (t) {
            console.warn("Local fonts API:", t);
        }
    }

    _getLocalFont({font: e}) {
        try {
            navigator?.permissions?.query ? navigator.permissions.query({name: "local-fonts"}).then((t) => {
                t.state === "granted" && this._sendLocalFont(e);
            }) : this._sendLocalFont(e);
        } catch (t) {
            console.warn("Local fonts API:", t);
        }
    }

    _unbusy() {
        this._lastDemandTime ? this._demandRender(this._lastDemandTime) : this.busy = !1;
    }

    _handleRVFC(e, {mediaTime: t, width: s, height: a}) {
        if (this._destroyed)
            return null;
        this.busy ? this._lastDemandTime = {mediaTime: t, width: s, height: a} : (this.busy = !0, this._demandRender({
            mediaTime: t,
            width: s,
            height: a
        })), this._video.requestVideoFrameCallback(this._handleRVFC.bind(this));
    }

    _demandRender({mediaTime: e, width: t, height: s}) {
        this._lastDemandTime = null, (t !== this._videoWidth || s !== this._videoHeight) && (this._videoWidth = t, this._videoHeight = s, this.resize()), this.sendMessage("demand", {time: e + this.timeOffset});
    }

    // if we're using offscreen render, we can't use ctx filters, so we can't use a transfered canvas
    _detachOffscreen() {
        if (!this._offscreenRender || this._ctx)
            return null;
        this._canvas.remove(), this._createCanvas(), this._canvasctrl = this._canvas, this._ctx = this._canvasctrl.getContext("2d"), this.sendMessage("detachOffscreen"), this.busy = !1, this.resize(0, 0, 0, 0, !0);
    }

    // if the video or track changed, we need to re-attach the offscreen canvas
    _reAttachOffscreen() {
        if (!this._offscreenRender || !this._ctx)
            return null;
        this._canvas.remove(), this._createCanvas(), this._canvasctrl = this._canvas.transferControlToOffscreen(), this._ctx = !1, this.sendMessage("offscreenCanvas", null, [this._canvasctrl]), this.resize(0, 0, 0, 0, !0);
    }

    _updateColorSpace() {
        this._video.requestVideoFrameCallback(() => {
            try {
                const e = new VideoFrame(this._video);
                this._videoColorSpace = _[e.colorSpace.matrix], e.close(), this.sendMessage("getColorSpace");
            } catch (e) {
                console.warn(e);
            }
        });
    }

    /**
     * Veryify the color spaces for subtitles and videos, then apply filters to correct the color of subtitles.
     * @param  {Object} options
     * @param  {String} options.subtitleColorSpace Subtitle color space. One of: BT601 BT709 SMPTE240M FCC
     * @param  {String=} options.videoColorSpace Video color space. One of: BT601 BT709
     */
    _verifyColorSpace({subtitleColorSpace: e, videoColorSpace: t = this._videoColorSpace}) {
        !e || !t || e !== t && (this._detachOffscreen(), this._ctx.filter = `url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg'><filter id='f'><feColorMatrix type='matrix' values='${f[e][t]} 0 0 0 0 0 1 0'/></filter></svg>#f")`);
    }

    _render({images: e, asyncRender: t, times: s, width: a, height: n, colorSpace: r}) {
        this._unbusy(), this.debug && (s.IPCTime = Date.now() - s.JSRenderTime), (this._canvasctrl.width !== a || this._canvasctrl.height !== n) && (this._canvasctrl.width = a, this._canvasctrl.height = n, this._verifyColorSpace({subtitleColorSpace: r})), this._ctx.clearRect(0, 0, this._canvasctrl.width, this._canvasctrl.height);
        for (const i of e)
            i.image && (t ? (this._ctx.drawImage(i.image, i.x, i.y), i.image.close()) : (this._bufferCanvas.width = i.w, this._bufferCanvas.height = i.h, this._bufferCtx.putImageData(new ImageData(this._fixAlpha(new Uint8ClampedArray(i.image)), i.w, i.h), 0, 0), this._ctx.drawImage(this._bufferCanvas, i.x, i.y)));
        if (this.debug) {
            s.JSRenderTime = Date.now() - s.JSRenderTime - s.IPCTime;
            let i = 0;
            const o = s.bitmaps || e.length;
            delete s.bitmaps;
            for (const d in s)
                i += s[d];
            console.log("Bitmaps: " + o + " Total: " + (i | 0) + "ms", s);
        }
    }

    _fixAlpha(e) {
        if (h._hasAlphaBug)
            for (let t = 3; t < e.length; t += 4)
                e[t] = e[t] > 1 ? e[t] : 1;
        return e;
    }

    _ready() {
        this._init(), this.dispatchEvent(new CustomEvent("ready"));
    }

    /**
     * Send data and execute function in the worker.
     * @param  {String} target Target function.
     * @param  {Object} [data] Data for function.
     * @param  {Transferable[]} [transferable] Array of transferables.
     */
    async sendMessage(e, t = {}, s) {
        await this._loaded, s ? this._worker.postMessage({
            target: e,
            transferable: s,
            ...t
        }, [...s]) : this._worker.postMessage({
            target: e,
            ...t
        });
    }

    _fetchFromWorker(e, t) {
        try {
            const s = e.target, a = setTimeout(() => {
                r(new Error("Error: Timeout while try to fetch " + s));
            }, 5e3), n = ({data: i}) => {
                i.target === s && (t(null, i), this._worker.removeEventListener("message", n), this._worker.removeEventListener("error", r), clearTimeout(a));
            }, r = (i) => {
                t(i), this._worker.removeEventListener("message", n), this._worker.removeEventListener("error", r), clearTimeout(a);
            };
            this._worker.addEventListener("message", n), this._worker.addEventListener("error", r), this._worker.postMessage(e);
        } catch (s) {
            this._error(s);
        }
    }

    _console({content: e, command: t}) {
        console[t].apply(console, JSON.parse(e));
    }

    _onmessage({data: e}) {
        this["_" + e.target] && this["_" + e.target](e);
    }

    _error(e) {
        const t = e instanceof Error ? e : e instanceof ErrorEvent ? e.error : new Error(e),
            s = e instanceof Event ? new ErrorEvent(e.type, e) : new ErrorEvent("error", {error: t});
        return this.dispatchEvent(s), console.error(t), t;
    }

    _removeListeners() {
        this._video && (this._ro && this._ro.unobserve(this._video), this._ctx && (this._ctx.filter = "none"), this._video.removeEventListener("timeupdate", this._boundTimeUpdate), this._video.removeEventListener("progress", this._boundTimeUpdate), this._video.removeEventListener("waiting", this._boundTimeUpdate), this._video.removeEventListener("seeking", this._boundTimeUpdate), this._video.removeEventListener("playing", this._boundTimeUpdate), this._video.removeEventListener("ratechange", this._boundSetRate), this._video.removeEventListener("resize", this._boundResize), this._video.removeEventListener("loadedmetadata", this._boundUpdateColorSpace));
    }

    /**
     * Destroy the object, worker, listeners and all data.
     * @param  {String|Error} [err] Error to throw when destroying.
     */
    destroy(e) {
        return e && (e = this._error(e)), this._video && this._canvasParent && this._video.parentNode?.removeChild(this._canvasParent), this._destroyed = !0, this._removeListeners(), this.sendMessage("destroy"), this._worker?.terminate(), e;
    }
}

export {
    h as default
};
