# shelfPlayer
 
A terminal cassette player for focused work.
 
![demo](demo.gif)
 
No album art. No recommendations. No feed. Just music and spinning reels in the corner of your screen while you work.
 
## what it is
 
shelfPlayer is a TUI music player built around the idea of going back into the cassette tape era. 
 
The idea came from wanting music while coding without the distraction of Spotify or Apple Music.


It also has a built-in Pomodoro timer. Enable focus mode when loading a tape and the player automatically pauses on breaks and resumes when your work session starts.
 
## features
 
- Spinning cassette reels rendered with polar coordinate math
- Local MP3 playback via beep
- Tape library — browse and manage your tapes from a shelf view
- Tape creator — search YouTube and download tracks directly into the app
- Auto-advance — plays through a tape start to finish
- Loop mode — repeat a track indefinitely (great for rain sounds or ambient music)
- Focus mode — Pomodoro timer baked into the player, music tied to work/break cycles
- Reels slow during breaks
## install
 
**Dependencies:**
 
```bash
# Go 1.21+
brew install go
 
# yt-dlp for downloading tracks
brew install yt-dlp
```
 
**Build from source:**
 
```bash
git clone https://github.com/rgarcia2304/shelfPlayer
cd shelfPlayer
make install
```
 
Then run from anywhere:
 
```bash
shelfplayer
```
 
## usage
 
### the shelf
 
When you open shelfPlayer you land on the shelf — a list of your tapes.
 
```
↑↓ . navigate    enter . inspect    n . new tape    q . quit
```
 
### tape detail
 
Select a tape to see its track list and choose how to listen.
 
- **just play** — loads the tape and plays straight through
- **focus** — starts a Pomodoro timer alongside the music. Use `+/-` to set the work session length before loading.
```
tab . switch mode    enter . play    esc . back
```
 
### the walkman
 
The main player screen. Two reels spin while music plays.
 
```
space . play/pause    n/p . tracks    l . loop    esc . shelf    q . quit
```
 
When focus mode is active the tape name row shows a countdown timer. The reels slow to a crawl during breaks and return to normal speed when your work session resumes.
 
### creating a tape
 
Hit `n` from the shelf to open the tape creator.
 
1. Name your tape
2. Search for a track — results come from YouTube via yt-dlp
3. Pick a result — it downloads in the background
4. Add more tracks or `ctrl+s` to save
Tapes are saved to `~/.shelfplayer/tapes/` as a folder of MP3s and a `tape.json` metadata file.
 
```
~/.shelfplayer/tapes/
  kind-of-blue/
    01-so-what.mp3
    02-freddie-freeloader.mp3
    tape.json
```
 
You can also add tapes manually — create a folder with MP3s and shelfPlayer will pick it up automatically on next launch, even without a `tape.json`.
 
## focus mode
 
Focus mode turns shelfPlayer into a Pomodoro timer tied to your music.
 
- Set your work session length on the tape detail screen (default 25 min)
- Music plays during work sessions
- Music pauses automatically on breaks
- Music resumes when the next work session starts
- After 4 sessions you get a longer 15 minute break
- The timer shows in the top right of the walkman display
The reels visually slow during breaks so you can tell at a glance what state you're in without reading the timer.
 
## manual tapes
 
You can build tapes without the creator by dropping MP3s into the tapes directory directly:
 
```bash
mkdir -p ~/.shelfplayer/tapes/my-tape
cp *.mp3 ~/.shelfplayer/tapes/my-tape/
```
 
shelfPlayer will scan the folder and build a tape automatically from whatever MP3s it finds. Add a `tape.json` for clean metadata:
 
```json
{
  "name": "My Tape",
  "artist": "Various",
  "color": "136",
  "tracks": [
    {
      "title": "Track Name",
      "artist": "Artist",
      "file": "01-track.mp3"
    }
  ]
}
```
 
## built with
 
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — terminal styling
- [beep](https://github.com/gopxl/beep) — audio playback
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) — YouTube audio download
## note on downloading
 
shelfPlayer uses yt-dlp to download audio from YouTube. Downloading from YouTube technically violates their Terms of Service. This tool is intended for personal use only. You are responsible for complying with the laws in your jurisdiction and the terms of any services you use.
 
---
 
cheers
