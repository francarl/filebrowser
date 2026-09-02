<template>
  <video ref="videoPlayer" class="video-max video-js" controls preload="auto">
    <source />
    <track
      kind="subtitles"
      v-for="(sub, index) in subtitles"
      :key="index"
      :src="sub"
      :label="subLabel(sub)"
      :default="index === 0"
    />
    <p class="vjs-no-js">
      Sorry, your browser doesn't support embedded videos, but don't worry, you
      can <a :href="source">download it</a>
      and watch it with your favorite video player!
    </p>
  </video>
</template>

<script setup lang="ts">
import { ref, onBeforeUnmount, nextTick } from "vue";
import videojs from "video.js";
import type Player from "video.js/dist/types/player";
import "videojs-mobile-ui";
import "videojs-hotkeys";
import "video.js/dist/video-js.min.css";
import "videojs-mobile-ui/dist/videojs-mobile-ui.css";

const videoPlayer = ref<HTMLElement | null>(null);
const player = ref<Player | null>(null);

const props = withDefaults(
  defineProps<{
    source: string;
    subtitles?: string[];
    options?: any;
  }>(),
  {
    options: {},
  }
);

const source = ref(props.source);
const sourceType = ref("");

const zoomrotate = {
  rotate: 0,
  zoom: 1,
  flipH: 1,
  flipV: 1,
  panX: 0,
  panY: 0,
};

const applyTransform = (el: HTMLElement) => {
  el.style.transform = `translate(${zoomrotate.panX || 0}px, ${zoomrotate.panY || 0}px) scale(${zoomrotate.zoom}) rotate(${zoomrotate.rotate}deg) scale(${zoomrotate.flipH}, ${zoomrotate.flipV})`;
};

const Button = videojs.getComponent("Button");

class RotateCustomButton extends Button {
  constructor(player: Player, options?: any) {
    super(player, options);
    // @ts-expect-error controlText is on Button prototype
    this.controlText("Rotate +90");
  }

  handleClick() {
    const vi = this.player().children()[0] as HTMLElement;
    zoomrotate.rotate += 90;
    applyTransform(vi);
  }

  override buildCSSClass() {
    return "vjs-icon-replay vjs-control vjs-button";
  }
}

class ZoomInCustomButton extends Button {
  constructor(player: Player, options?: any) {
    super(player, options);
    // @ts-expect-error controlText is on Button prototype
    this.controlText("Zoom in");
  }

  handleClick() {
    const vi = this.player().children()[0] as HTMLElement;
    zoomrotate.zoom += 0.1;
    applyTransform(vi);
  }

  override buildCSSClass() {
    return "vjs-icon-circle vjs-control vjs-button";
  }
}

class ZoomOutCustomButton extends Button {
  constructor(player: Player, options?: any) {
    super(player, options);
    // @ts-expect-error controlText is on Button prototype
    this.controlText("Zoom out");
  }

  handleClick() {
    const vi = this.player().children()[0] as HTMLElement;
    zoomrotate.zoom -= 0.1;
    applyTransform(vi);
  }

  override buildCSSClass() {
    return "vjs-icon-circle-outline vjs-control vjs-button";
  }
}

class FlipHCustomButton extends Button {
  constructor(player: Player, options?: any) {
    super(player, options);
    // @ts-expect-error controlText is on Button prototype
    this.controlText("Flip H");
  }

  handleClick() {
    const vi = this.player().children()[0] as HTMLElement;
    zoomrotate.flipH *= -1;
    applyTransform(vi);
  }

  override buildCSSClass() {
    return "vjs-icon-fliph vjs-control vjs-button";
  }
}

class FlipVCustomButton extends Button {
  constructor(player: Player, options?: any) {
    super(player, options);
    // @ts-expect-error controlText is on Button prototype
    this.controlText("Flip V");
  }

  handleClick() {
    const vi = this.player().children()[0] as HTMLElement;
    zoomrotate.flipV *= -1;
    applyTransform(vi);
  }

  override buildCSSClass() {
    return "vjs-icon-flipv vjs-control vjs-button";
  }
}

class DownloadFrameButton extends Button {
  constructor(player: Player, options?: any) {
    super(player, options);
  }

  override createEl() {
    const el = super.createEl("button", {
      className: "vjs-custom-button",
    });
    el.innerHTML = '<i class="material-icons">photo_camera</i>';
    return el;
  }

  handleClick() {
    const video = this.player().tech_.el() as HTMLVideoElement;
    const canvas = document.createElement("canvas");
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    const ctx = canvas.getContext("2d");
    ctx?.drawImage(video, 0, 0, canvas.width, canvas.height);
    const link = document.createElement("a");
    link.download = `frame-${this.player().currentTime()}.png`;
    link.href = canvas.toDataURL("image/png");
    link.click();
  }
}

class FrameBackButton extends Button {
  constructor(player: Player, options?: any) {
    super(player, options);
  }

  override createEl() {
    const el = super.createEl("button", {
      className: "vjs-custom-button",
    });
    el.innerHTML = '<span class="vjs-fbf">&lt;</span>';
    return el;
  }

  handleClick() {
    const p = this.player()!;
    p.pause();
    p.currentTime(p.currentTime()! - 1 / 60);
  }
}

class FrameForwardButton extends Button {
  constructor(player: Player, options?: any) {
    super(player, options);
  }

  override createEl() {
    const el = super.createEl("button", {
      className: "vjs-custom-button",
    });
    el.innerHTML = '<span class="vjs-fbf">&gt;</span>';
    return el;
  }

  handleClick() {
    const p = this.player()!;
    p.pause();
    p.currentTime(p.currentTime()! + 1 / 60);
  }
}

class ResetTransformButton extends Button {
  constructor(player: Player, options?: any) {
    super(player, options);
    // @ts-expect-error controlText is on Button prototype
    this.controlText("Reset transform");
  }

  handleClick() {
    zoomrotate.rotate = 0;
    zoomrotate.zoom = 1;
    zoomrotate.flipH = 1;
    zoomrotate.flipV = 1;
    zoomrotate.panX = 0;
    zoomrotate.panY = 0;
    const vi = this.player().children()[0] as HTMLElement;
    applyTransform(vi);
  }

  override buildCSSClass() {
    return "vjs-icon-replay vjs-control vjs-button";
  }
}

interface TouchState {
  startX: number;
  startY: number;
  startPanX: number;
  startPanY: number;
  startDist: number;
  startZoom: number;
  active: boolean;
}

let touchState: TouchState | null = null;

const getTouchDistance = (touches: TouchList) => {
  if (touches.length < 2) return 0;
  const dx = touches[0].clientX - touches[1].clientX;
  const dy = touches[0].clientY - touches[1].clientY;
  return Math.sqrt(dx * dx + dy * dy);
};

const getTouchCenter = (touches: TouchList) => {
  return {
    x: (touches[0].clientX + touches[1].clientX) / 2,
    y: (touches[0].clientY + touches[1].clientY) / 2,
  };
};

const setupTouchListeners = (techEl: HTMLElement) => {
  techEl.addEventListener('touchstart', (e) => {
    if (e.touches.length === 2) {
      e.preventDefault();
      const center = getTouchCenter(e.touches);
      touchState = {
        startX: center.x,
        startY: center.y,
        startPanX: zoomrotate.panX,
        startPanY: zoomrotate.panY,
        startDist: getTouchDistance(e.touches),
        startZoom: zoomrotate.zoom,
        active: true,
      };
    } else if (e.touches.length === 1) {
      touchState = {
        startX: e.touches[0].clientX,
        startY: e.touches[0].clientY,
        startPanX: zoomrotate.panX,
        startPanY: zoomrotate.panY,
        startDist: 0,
        startZoom: zoomrotate.zoom,
        active: true,
      };
    }
  }, { passive: false });

  techEl.addEventListener('touchmove', (e) => {
    if (!touchState?.active) return;
    e.preventDefault();

    if (e.touches.length === 2) {
      const dist = getTouchDistance(e.touches);
      const center = getTouchCenter(e.touches);
      const scale = dist / touchState.startDist;
      const newZoom = Math.max(0.5, Math.min(5, touchState.startZoom * scale));
      const deltaPanX = center.x - touchState.startX;
      const deltaPanY = center.y - touchState.startY;
      zoomrotate.zoom = newZoom;
      zoomrotate.panX = touchState.startPanX + deltaPanX * 0.5;
      zoomrotate.panY = touchState.startPanY + deltaPanY * 0.5;
      const vi = player.value!.children()[0] as HTMLElement;
      applyTransform(vi);
    } else if (e.touches.length === 1) {
      const deltaPanX = e.touches[0].clientX - touchState.startX;
      const deltaPanY = e.touches[0].clientY - touchState.startY;
      zoomrotate.panX = touchState.startPanX + deltaPanX * 0.5;
      zoomrotate.panY = touchState.startPanY + deltaPanY * 0.5;
      const vi = player.value!.children()[0] as HTMLElement;
      applyTransform(vi);
    }
  }, { passive: false });

  techEl.addEventListener('touchend', () => {
    touchState = null;
  });

  techEl.addEventListener('wheel', (e) => {
    if (e.ctrlKey) {
      e.preventDefault();
      const delta = -e.deltaY * 0.001;
      const newZoom = Math.max(0.5, Math.min(5, zoomrotate.zoom + delta * zoomrotate.zoom));
      zoomrotate.zoom = newZoom;
      const vi = player.value!.children()[0] as HTMLElement;
      applyTransform(vi);
    }
  }, { passive: false });
};

videojs.registerComponent("rotateCustomButton", RotateCustomButton);
videojs.registerComponent("zoomInCustomButton", ZoomInCustomButton);
videojs.registerComponent("zoomOutCustomButton", ZoomOutCustomButton);
videojs.registerComponent("flipHCustomButton", FlipHCustomButton);
videojs.registerComponent("flipVCustomButton", FlipVCustomButton);
videojs.registerComponent("downloadFrameButton", DownloadFrameButton);
videojs.registerComponent("frameBackButton", FrameBackButton);
videojs.registerComponent("frameForwardButton", FrameForwardButton);
videojs.registerComponent("resetTransformButton", ResetTransformButton);

nextTick(() => {
  initVideoPlayer();
});

onBeforeUnmount(() => {
  if (player.value) {
    player.value.dispose();
    player.value = null;
  }
});

const initVideoPlayer = async () => {
  try {
    const lang = document.documentElement.lang;
    const languagePack = await (
      languageImports[lang] || languageImports.en
    )?.();
    const code = languageImports[lang] ? lang : "en";
    videojs.addLanguage(code, languagePack.default);
    sourceType.value = "";

    sourceType.value = getSourceType(source.value);

    const srcOpt = { sources: { src: props.source, type: sourceType.value } };
    const langOpt = { language: code };
    const playbackRatesOpt = { playbackRates: [0.1, 0.2, 0.3, 0.4, 0.5, 0.8, 1.0, 1.5, 2.0, 4.0, 8.0] };
    const options = getOptions(
      props.options,
      langOpt,
      srcOpt,
      playbackRatesOpt
    );
    player.value = videojs(videoPlayer.value!, options, () => {
      onPlayerReady();
    });
  } catch (error) {
    console.error("Error initializing video player:", error);
  }
};

const onPlayerReady = () => {
  const controlBar = player.value!.getChild("ControlBar");
  if (controlBar) {
    controlBar.addChild("rotateCustomButton", {});
    controlBar.addChild("zoomInCustomButton", {});
    controlBar.addChild("zoomOutCustomButton", {});
    controlBar.addChild("flipHCustomButton", {});
    controlBar.addChild("flipVCustomButton", {});
    controlBar.addChild("downloadFrameButton", {});
    controlBar.addChild("frameBackButton", {});
    controlBar.addChild("frameForwardButton", {});
    controlBar.addChild("resetTransformButton", {});
  }

  const techEl = player.value!.tech({ el: () => videoPlayer.value! })?.el() as HTMLElement | undefined;
  if (techEl) {
    setupTouchListeners(techEl);
  }

  player.value!.on('loadedmetadata', () => {
    const videoTech = player.value!.tech({ el: () => videoPlayer.value! })?.el() as HTMLVideoElement;
    if (videoTech) {
      if (videoTech.videoHeight > videoTech.videoWidth) {
        zoomrotate.zoom = videoTech.videoHeight / videoTech.videoWidth;
        zoomrotate.rotate = -90;
      } else {
        zoomrotate.zoom = 1;
        zoomrotate.rotate = 0;
      }
      applyTransform(videoTech);
    }
  });

  // TODO: need to test on mobile
  // @ts-expect-error mobileUi is a custom plugin
  player.value!.mobileUi();
};

const getOptions = (...srcOpt: any[]) => {
  const options = {
    controlBar: {
      skipButtons: false,
      pictureInPictureToggle: false,
    },
    html5: {
      nativeTextTracks: false,
    },
    plugins: {
      hotkeys: {
        volumeStep: 0.1,
        seekStep: 10,
        enableModifiersForNumbers: false,
        shortcuts: {
          'alt+r': () => {
            const vi = player.value!.children()[0] as HTMLElement;
            zoomrotate.rotate += 90;
            applyTransform(vi);
          },
          'alt+z': () => {
            const vi = player.value!.children()[0] as HTMLElement;
            zoomrotate.zoom += 0.1;
            applyTransform(vi);
          },
          'alt+x': () => {
            const vi = player.value!.children()[0] as HTMLElement;
            zoomrotate.zoom -= 0.1;
            applyTransform(vi);
          },
          'n': () => {
            const ct = player.value!.currentTime() || 0;
            player.value!.currentTime(ct - 10);
            player.value!.play();
          },
          'm': () => {
            const ct = player.value!.currentTime() || 0;
            player.value!.currentTime(ct + 10);
            player.value!.play();
          },
          'k': () => {
            player.value!.pause();
            const ct = player.value!.currentTime() || 0;
            player.value!.currentTime(ct - 1 / 60);
          },
          'l': () => {
            player.value!.pause();
            const ct = player.value!.currentTime() || 0;
            player.value!.currentTime(ct + 1 / 60);
          },
        },
      },
      muted: true,
    },
  };

  return videojs.obj.merge(options, ...srcOpt);
};

const getSourceType = (source: string) => {
  const fileExtension = source ? source.split("?")[0].split(".").pop() : "";
  if (fileExtension?.toLowerCase() === "mkv") {
    return "video/mp4";
  }
  return "";
};

const subLabel = (subUrl: string) => {
  let url: URL;
  try {
    url = new URL(subUrl);
  } catch {
    url = new URL(subUrl, window.location.origin);
  }

  const label = decodeURIComponent(
    url.pathname
      .split("/")
      .pop()!
      .replace(/\.[^/.]+$/, "")
  );

  return label;
};

interface LanguageImports {
  [key: string]: () => Promise<any>;
}

const languageImports: LanguageImports = {
  ar: () => import("video.js/dist/lang/ar.json"),
  bg: () => import("video.js/dist/lang/bg.json"),
  cs: () => import("video.js/dist/lang/cs.json"),
  de: () => import("video.js/dist/lang/de.json"),
  el: () => import("video.js/dist/lang/el.json"),
  en: () => import("video.js/dist/lang/en.json"),
  es: () => import("video.js/dist/lang/es.json"),
  fr: () => import("video.js/dist/lang/fr.json"),
  he: () => import("video.js/dist/lang/he.json"),
  hr: () => import("video.js/dist/lang/hr.json"),
  hu: () => import("video.js/dist/lang/hu.json"),
  it: () => import("video.js/dist/lang/it.json"),
  ja: () => import("video.js/dist/lang/ja.json"),
  ko: () => import("video.js/dist/lang/ko.json"),
  lv: () => import("video.js/dist/lang/lv.json"),
  nb: () => import("video.js/dist/lang/nb.json"),
  nl: () => import("video.js/dist/lang/nl.json"),
  "nl-be": () => import("video.js/dist/lang/nl.json"),
  pl: () => import("video.js/dist/lang/pl.json"),
  "pt-br": () => import("video.js/dist/lang/pt-BR.json"),
  "pt-pt": () => import("video.js/dist/lang/pt-PT.json"),
  ro: () => import("video.js/dist/lang/ro.json"),
  ru: () => import("video.js/dist/lang/ru.json"),
  sk: () => import("video.js/dist/lang/sk.json"),
  tr: () => import("video.js/dist/lang/tr.json"),
  uk: () => import("video.js/dist/lang/uk.json"),
  vi: () => import("video.js/dist/lang/vi.json"),
  "zh-cn": () => import("video.js/dist/lang/zh-CN.json"),
  "zh-tw": () => import("video.js/dist/lang/zh-TW.json"),
};
</script>
<style scoped>
.video-max {
  width: 100%;
  height: 100%;
}
</style>
<style>
.vjs-icon-fliph:before { content: '\21D4'; font-size: 1.3em; }
.vjs-icon-flipv:before { content: '\21D5'; font-size: 1.3em; }
.vjs-fbf {
  border: 1px solid white;
  padding: 2px 3px;
  border-radius: 2px;
}
</style>
