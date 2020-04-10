## Checklist for adding a new provider
1. Create a sample yaml cluster manifest with declaration for possible options in yaml and its naming.  
   Naming for the options should be aligned with correspondent names in terraform provider or module.
2. Create a sample Github Action and Gitlab pipeline defining required cloud authentication credentials, like Username/Password, Personal Access Tokens, or Access files, ex:
```yaml
    # For user and password please use your token name and token hash
    # https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/
        cloud-user: ${{ secrets.DO_TOKEN_NAME }}
        cloud-pass: ${{ secrets.DIGITALOCEAN_TOKEN }}
```
3. Define required treatment in main function:
```yaml
# entrypoint.sh
    digitalocean)
        DEBUG "Cloud Provider: Digital Ocean"
        ;;
```
4. Add a new state storage creation function.
