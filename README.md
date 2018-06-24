# iijmio-checker

This is a tiny program to checker amount of your use for IIJmio SIM.

## Usage

1. Build the binary.

    ```bash
    # build `iijmio-checker` binary
    make build
    ```

2. Run the server to generate token for IIJmio API.

    ```bash
    # Open http://localhost:8080 and follow the instruction.
    ./iijmio-checker auth
    ```

3. Run the job.

    ```sh
    ./iijmio-checker cron
    ```

  It will show the result if your amount is above the threshold. If you prefer,
  you can set this to crontab to check & mail the result.
