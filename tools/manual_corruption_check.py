#!/usr/bin/env python3
"""
Media file corruption detector using ffprobe.
Recursively scans a directory for corrupted media files and writes results to corrupt.txt
"""

import os
import sys
import subprocess
from pathlib import Path

# Common media file extensions
MEDIA_EXTENSIONS = {
    '.mp4', '.avi', '.mkv', '.mov', '.wmv', '.flv', '.webm', '.m4v',
    '.mp3', '.wav', '.flac', '.aac', '.ogg', '.wma', '.m4a',
    '.jpg', '.jpeg', '.png', '.gif', '.bmp', '.tiff', '.webp'
}

def is_media_file(file_path):
    """Check if file has a media extension"""
    return file_path.suffix.lower() in MEDIA_EXTENSIONS

def check_corruption(file_path):
    """
    Check if a media file is corrupted using ffmpeg decode test
    Returns: (is_corrupted: bool, error_message: str)
    """
    try:
        # Use ffmpeg to actually decode the entire file to null
        # This catches corruption that ffprobe misses
        cmd = [
            'ffmpeg',
            '-v', 'error',
            '-i', str(file_path),
            '-f', 'null',
            '-'
        ]
        
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=60  # Longer timeout since we're decoding
        )
        
        # Check return code
        if result.returncode != 0:
            error_msg = result.stderr.strip() if result.stderr else "ffmpeg decode failed"
            return True, error_msg
        
        # Check stderr for corruption indicators even if return code is 0
        stderr_text = result.stderr.strip()
        if stderr_text:
            # Look for common corruption indicators
            corruption_keywords = [
                'error', 'corrupt', 'invalid', 'broken', 'truncated', 
                'damaged', 'failed', 'missing', 'decode error', 
                'header damaged', 'no frame', 'stream', 'unexpected'
            ]
            
            stderr_lower = stderr_text.lower()
            found_issues = [kw for kw in corruption_keywords if kw in stderr_lower]
            
            if found_issues:
                return True, f"decode issues: {stderr_text[:200]}..."
        
        return False, ""
        
    except subprocess.TimeoutExpired:
        return True, "ffmpeg timeout - file may be corrupted or very large"
    except FileNotFoundError:
        return True, "ffmpeg not found - please install ffmpeg"
    except Exception as e:
        return True, f"unexpected error: {str(e)}"

def get_directory():
    """Prompt user for directory path"""
    while True:
        directory = input("Enter directory path to scan: ").strip()
        
        # Handle quotes around path
        if directory.startswith('"') and directory.endswith('"'):
            directory = directory[1:-1]
        elif directory.startswith("'") and directory.endswith("'"):
            directory = directory[1:-1]
        
        # Expand user path (~)
        directory = os.path.expanduser(directory)
        
        if not directory:
            print("Please enter a directory path.")
            continue
            
        if not os.path.exists(directory):
            print(f"Directory '{directory}' does not exist. Please try again.")
            continue
            
        if not os.path.isdir(directory):
            print(f"'{directory}' is not a directory. Please try again.")
            continue
            
        return directory

def get_verbose_choice():
    """Ask user if they want verbose output"""
    while True:
        choice = input("Show detailed progress for each file? (y/n): ").strip().lower()
        if choice in ['y', 'yes']:
            return True
        elif choice in ['n', 'no']:
            return False
        else:
            print("Please enter 'y' for yes or 'n' for no.")

def scan_directory(directory_path, verbose=False):
    """
    Recursively scan directory for corrupted media files
    Returns list of corrupted files
    """
    directory = Path(directory_path)
    
    corrupted_files = []
    total_files = 0
    
    print(f"Scanning directory: {directory_path}")
    
    # Walk through all files recursively
    for file_path in directory.rglob('*'):
        if file_path.is_file() and is_media_file(file_path):
            total_files += 1
            
            if verbose:
                print(f"Checking: {file_path}")
            
            is_corrupted, error_msg = check_corruption(file_path)
            
            if is_corrupted:
                corrupted_files.append((str(file_path), error_msg))
                print(f"CORRUPTED: {file_path}")
                if verbose and error_msg:
                    print(f"  Error: {error_msg}")
            elif verbose:
                print(f"OK: {file_path}")
    
    print(f"\nScan complete: {total_files} media files checked")
    print(f"Found {len(corrupted_files)} corrupted files")
    
    return corrupted_files

def write_results(directory_path, corrupted_files):
    """Write corrupted files list to corrupt.txt"""
    output_file = Path(directory_path) / "corrupt.txt"
    
    try:
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(f"Corrupted Media Files Report\n")
            f.write(f"Generated by corruption detector\n")
            f.write(f"Directory: {directory_path}\n")
            f.write(f"Total corrupted files: {len(corrupted_files)}\n")
            f.write("=" * 50 + "\n\n")
            
            if corrupted_files:
                for file_path, error_msg in corrupted_files:
                    f.write(f"{file_path}\n")
                    if error_msg:
                        f.write(f"  Error: {error_msg}\n")
                    f.write("\n")
            else:
                f.write("No corrupted files found!\n")
        
        print(f"Results written to: {output_file}")
        
    except Exception as e:
        print(f"Error writing results file: {e}")

def main():
    print("Media File Corruption Detector")
    print("=" * 35)
    
    # Check if ffprobe is available
    try:
        subprocess.run(['ffprobe', '-version'], capture_output=True, check=True)
    except (subprocess.CalledProcessError, FileNotFoundError):
        print("Error: ffprobe not found. Please install ffmpeg.")
        sys.exit(1)
    
    # Get directory from user
    directory = get_directory()
    
    # Get verbose preference
    verbose = get_verbose_choice()
    
    print()  # Add some spacing
    
    # Scan directory
    corrupted_files = scan_directory(directory, verbose)
    
    # Write results
    write_results(directory, corrupted_files)

if __name__ == "__main__":
    main()