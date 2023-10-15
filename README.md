# video uploader

## Use-cases

- Sign up/Sign in
- Upload a video
- List videos

## How to process an upload as a buffer

1 - Create a multipart reader from the request
2 - Read the parts available from the multipart reader
3 - In a loop read each part into its own buffer   

## File upload requirements

1 - Each request can hold only one file
2 - A file can be at most 100MB large
3 - A file can be of any extension
4 - Upload