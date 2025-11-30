# Video Downloader

User-friendly tool to download full videos or clips from [1000+ websites](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md), powered by [yt-dlp](https://github.com/yt-dlp/yt-dlp).

## Table of Contents

- [How to Use](#how-to-use)
- [What Happens When You Run](#what-happens-when-you-run)
- [How to Format URLs](#how-to-format-urls)
- [Custom Download Location](#custom-download-location)
- [Clip Modes](#clip-modes)
- [Demo](#demo)

## How to Use

### Step 1: Download
Go to [Releases](https://github.com/mohamedhesham1619/video-downloader/releases) and download the right version for your system

Supported systems:

- Windows (fully tested) ✓
- macOS (untested)
- Linux (untested)

> Note: Only the Windows version has been thoroughly tested. macOS and Linux versions should work but haven't been verified yet. If you encounter issues, please report them.

### Step 2: Extract
Unzip the downloaded file to a location of your choice.

### Step 3: Add Video Links
Open `urls.txt` in a text editor and add your video URLs (one per line), then save the file. Check [How to Format URLs](#how-to-format-urls) for examples.

### Step 4: Run
- **Windows:** Double-click `downloader.exe`
- **macOS/Linux:** Open Terminal in the app folder and run `./downloader`

### Step 5: Done!
Your videos will be saved in the `Downloads` folder inside the app folder.

## What Happens When You Run

The app automatically manages its dependencies:

**First time you run:**
- Checks for required tools (yt-dlp, ffmpeg, deno) in the `bin` folder inside the app folder
- Downloads any missing tools
- This may take a moment depending on your internet speed

**Every time after:**
- Checks if there's a newer version of yt-dlp available
- Downloads the update if found

You don't need to install anything manually - the app handles everything for you.

> Note: If you already have these tools, you can create a `bin` folder inside the app folder and place them there to save download time.

## How to Format URLs

Each line must start with the URL, optionally followed by quality or time range.

**Behavior:**
- No time range → downloads the full video
- With time range → downloads only that part
- No quality → downloads best available quality
- With quality → uses the specified quality

**Formats:**
- Quality: Any number with "p" (e.g., `360p`, `720p`, `1080p`, `2160p`)
- Time range: `HH:MM:SS-HH:MM:SS`

**Examples:**
```
# Downloads full video in best available quality
https://youtube.com/watch?v=example

# Downloads full video in 720p quality
https://youtube.com/watch?v=example 720p

# Downloads only the clip from 1:30 to 2:45 in best available quality
https://youtube.com/watch?v=example 00:01:30-00:02:45

# Downloads clip from 1:30 to 2:45 in 720p quality
https://youtube.com/watch?v=example 720p 00:01:30-00:02:45

# Downloads clip from 1:30 to 2:45 in 1080p quality (the order after the URL doesn't matter)
https://youtube.com/watch?v=example 00:01:30-00:02:45 1080p
```

## Custom Download Location

By default, videos are saved to the `Downloads` folder inside the app folder. To save to a different location:

1. Open Terminal/Command Prompt in the app folder
   - **Windows:** Shift + Right-click in the folder → "Open PowerShell window here" or "Open Command Prompt here"
   - **macOS/Linux:** Right-click in the folder → "Open Terminal here"

2. Run the app with the `-path` flag and the full path to your desired location:

**Windows:**
```
./downloader.exe -path "D:\My Videos"
```

**macOS/Linux:**
```
./downloader -path "/home/user/Videos"
```

## Clip Modes

When downloading clips, you'll be asked to choose a mode:

**Fast Mode (Recommended)**
- Simply copies the video without re-processing it
- Much faster, especially for long clips
- May start a few seconds early or have a brief freeze at the beginning
- Best for most use cases

**Accurate Mode**
- Re-processes the video to cut at exact times
- Cuts clips very precisely at the exact times you specify
- Speed depends on your computer's hardware


> Tip: Always start with Fast mode and only switch to Accurate mode if you notice timing issues or frozen frames in your clips.

The app automatically tries to use your graphics card (GPU) first for faster processing in Accurate mode, and falls back to your CPU if the GPU isn't available. If you see a message about "falling back to CPU encoder," try updating your graphics card drivers for better performance.


## Demo





