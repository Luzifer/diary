# Luzifer / diary

`diary` is a small utility to write a personal, encrypted markdown diary. It will prepend new entries using a template so that you don't have to write the skeleton yourself. The diary stored on the disk is encrypted using AES-256 encyption. This ensures your data is secure as long as nobody gets your password.

## FAQ

1. **When working with `vim` the filetype is not detected, how to solve this?**  
   This is because of the file contents are copied to a temporary file for editing. You can solve this by a line at the end of your diary with this content: `<!-- vim: ft=markdown -->`
