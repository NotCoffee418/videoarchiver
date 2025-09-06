# Install https://github.com/acoustid/chromaprint
#
# pip install colorama
#
# Add to env (ps1):
# $env:Path += ";C:\chromaprint"
# [Environment]::SetEnvironmentVariable("Path", $env:Path, "User")
# (restart terminal)
#
# Usage:
# python tools/manual_duplicate_check.py

import os
import subprocess
import sys
import shutil
import json
from itertools import combinations
from colorama import Fore, Style, init
from datetime import datetime

init(autoreset=True)

# -------------------------------
# CONFIG
AUDIO_EXTENSIONS = (".mp3", ".flac", ".wav", ".aac", ".ogg", ".m4a")
SIMILARITY_THRESHOLD = 0.90  # 90%
FPCALC_LENGTH = 120  # seconds of audio to fingerprint
SAVE_INTERVAL = 50  # save cache every 50 new fingerprints
# -------------------------------

def check_fpcalc():
    if not shutil.which("fpcalc"):
        print(f"{Fore.RED}[ERROR]{Style.RESET_ALL} 'fpcalc' not found. Install from https://acoustid.org/chromaprint")
        sys.exit(1)

def get_fingerprint(file_path):
    if not os.path.isfile(file_path):
        print(f"{Fore.RED}[ERROR]{Style.RESET_ALL} File not found: {file_path}")
        return None, None
    try:
        result = subprocess.run(
            ["fpcalc", "-raw", "-length", str(FPCALC_LENGTH), file_path],
            capture_output=True, text=True, check=True
        )
        lines = result.stdout.splitlines()
        duration = None
        fingerprint = None

        for line in lines:
            if line.startswith("DURATION="):
                duration = int(line.split("=")[1])
            elif line.startswith("FINGERPRINT="):
                fingerprint = line.split("=")[1]

        if duration is None or fingerprint is None:
            print(f"{Fore.YELLOW}[WARN]{Style.RESET_ALL} Incomplete fingerprint data for: {file_path}")
            return None, None

        return duration, fingerprint

    except subprocess.CalledProcessError as e:
        print(f"{Fore.RED}[ERROR]{Style.RESET_ALL} Failed to fingerprint: {file_path}")
        print(f"{Fore.RED}[STDERR]{Style.RESET_ALL} {e.stderr.strip()}")
        return None, None

def fingerprint_similarity(fp1, fp2):
    tokens1 = set(fp1.split(','))
    tokens2 = set(fp2.split(','))
    common = len(tokens1 & tokens2)
    total = max(len(tokens1), len(tokens2))
    return common / total if total > 0 else 0

def scan_audio_files(folder):
    print(f"{Fore.CYAN}[SCAN]{Style.RESET_ALL} Searching folder: {folder}")
    audio_files = []

    for root, _, files in os.walk(folder):
        for name in files:
            if name.lower().endswith(AUDIO_EXTENSIONS):
                full_path = os.path.join(root, name)
                audio_files.append(full_path)

    return audio_files

def load_fingerprint_cache(cache_path):
    if os.path.isfile(cache_path):
        try:
            with open(cache_path, "r", encoding="utf-8") as f:
                return json.load(f)
        except Exception as e:
            print(f"{Fore.YELLOW}[WARN]{Style.RESET_ALL} Failed to load cache: {e}")
    return {}

def save_fingerprint_cache(cache_path, cache_data):
    try:
        with open(cache_path, "w", encoding="utf-8") as f:
            json.dump(cache_data, f, indent=2)
    except Exception as e:
        print(f"{Fore.RED}[ERROR]{Style.RESET_ALL} Failed to save cache: {e}")

def clean_cache(fp_cache):
    print(f"{Fore.CYAN}[INFO]{Style.RESET_ALL} Checking for missing files in fingerprint cache...")
    original_len = len(fp_cache)
    removed = 0
    new_cache = {}

    for path, data in fp_cache.items():
        if os.path.isfile(path):
            new_cache[path] = data
        else:
            print(f"{Fore.YELLOW}[REMOVED]{Style.RESET_ALL} File missing, removed from cache: {path}")
            removed += 1

    if removed > 0:
        print(f"{Fore.CYAN}[CLEAN]{Style.RESET_ALL} Removed {removed} missing file(s) from cache")
    return new_cache

def main():
    check_fpcalc()

    print(f"{Fore.CYAN}== Audio Duplicate Finder =={Style.RESET_ALL}")
    folder = input("Enter path to music folder: ").strip('"')

    if not os.path.isdir(folder):
        print(f"{Fore.RED}[ERROR]{Style.RESET_ALL} Not a valid folder.")
        return

    cache_path = os.path.join(folder, "fp.json")
    log_file = os.path.join(folder, "dups.txt")

    fp_cache = load_fingerprint_cache(cache_path)
    fp_cache = clean_cache(fp_cache)

    files = scan_audio_files(folder)
    print(f"{Fore.GREEN}[INFO]{Style.RESET_ALL} Found {len(files)} audio files.\n")

    new_fingerprints = 0

    for idx, file_path in enumerate(files, 1):
        if file_path not in fp_cache:
            dur, fp = get_fingerprint(file_path)
            if fp:
                fp_cache[file_path] = [dur, fp]  # store as list for JSON
                msg = f"{Fore.GREEN}[OK]{Style.RESET_ALL} Fingerprinted: {file_path}"
                new_fingerprints += 1
            else:
                msg = f"{Fore.YELLOW}[SKIP]{Style.RESET_ALL} Could not fingerprint: {file_path}"
        else:
            msg = f"{Fore.BLUE}[CACHED]{Style.RESET_ALL} Using cached fingerprint: {file_path}"

        print(msg)

        if new_fingerprints > 0 and new_fingerprints % SAVE_INTERVAL == 0:
            save_fingerprint_cache(cache_path, fp_cache)
            print(f"{Fore.CYAN}[INFO]{Style.RESET_ALL} Saved fingerprint cache after {new_fingerprints} new files")

    if new_fingerprints % SAVE_INTERVAL != 0:
        save_fingerprint_cache(cache_path, fp_cache)
        print(f"{Fore.CYAN}[INFO]{Style.RESET_ALL} Saved fingerprint cache (final)")

    print(f"\n{Fore.CYAN}== Scanning for duplicates... =={Style.RESET_ALL}\n")

    with open(log_file, "w", encoding="utf-8") as logf:
        logf.write(f"# Audio Duplicate Scan - {datetime.now()}\n")
        logf.write(f"# Folder: {folder}\n\n")

    for file1, file2 in combinations(fp_cache.keys(), 2):
        dur1, fp1 = fp_cache[file1]
        dur2, fp2 = fp_cache[file2]

        similarity = fingerprint_similarity(fp1, fp2)

        if similarity >= SIMILARITY_THRESHOLD:
            percent = round(similarity * 100, 2)
            msg = f"[MATCH {percent}%]\n  ├── {file1}\n  └── {file2}\n"
            print(f"{Fore.MAGENTA}[MATCH {percent}%]{Style.RESET_ALL}")
            print(f"  ├── {file1}")
            print(f"  └── {file2}\n")

            with open(log_file, "a", encoding="utf-8") as logf:
                logf.write(msg)

    print(f"{Fore.CYAN}[INFO]{Style.RESET_ALL} Duplicate scan complete. Results saved to {log_file}")

if __name__ == "__main__":
    main()
