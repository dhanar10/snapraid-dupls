# snapraid-dupls

A tool for helping to take care duplicate files listed by `snapraid dup`.

## Usage Example

1. Get a `snapraid dup` log file.

    ```sh
    snapraid -l snapraid-dup.log dup
    ```

2. Get a list of duplicate files that may be deleted with optional path regex and minimum size filter.

    ```sh
    snapraid-dupls -regex '^/storage/' -minbytes 1000000 snapraid-dup.log 
    ...
    /storage/data/photos/photos97.jpg
    /storage/data/photos/photos98.jpg
    /storage/data/photos/photos99.jpg
    # File path regex: ^/storage/
    # File minimum bytes: 1000000
    # Total files: 2050
    # Total bytes: 30647203029
    # Suggested delete command: snapraid-dupls -regex '.*' -minbytes 1000000 snapraid-dup.log | xargs -I{} rm -v '{}'
    ```

3. Execute the suggested delete command to delete the files.
