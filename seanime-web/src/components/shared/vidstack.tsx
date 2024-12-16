import { defaultLayoutIcons } from "@vidstack/react/player/layouts/default"
import { LuCast, LuVolume1, LuVolume2, LuVolumeX } from "react-icons/lu"
import {
    RiClosedCaptioningFill,
    RiClosedCaptioningLine,
    RiFullscreenExitLine,
    RiFullscreenLine,
    RiPauseLargeLine,
    RiPictureInPictureExitLine,
    RiPictureInPictureLine,
    RiPlayLargeLine,
    RiResetLeftFill,
    RiSettings4Line,
} from "react-icons/ri"

export const vidstackLayoutIcons = {
    ...defaultLayoutIcons,
    PlayButton: {
        Play: RiPlayLargeLine,
        Pause: RiPauseLargeLine,
        Replay: RiResetLeftFill,
    },
    MuteButton: {
        Mute: LuVolumeX,
        VolumeLow: LuVolume1,
        VolumeHigh: LuVolume2,
    },
    GoogleCastButton: {
        Default: LuCast,
    },
    PIPButton: {
        Enter: RiPictureInPictureLine,
        Exit: RiPictureInPictureExitLine,
    },
    FullscreenButton: {
        Enter: RiFullscreenLine,
        Exit: RiFullscreenExitLine,
    },
    Menu: {
        ...defaultLayoutIcons["Menu"],
        Settings: RiSettings4Line,
    },
    CaptionButton: {
        On: RiClosedCaptioningFill,
        Off: RiClosedCaptioningLine,
    },
}
