---
subcategory: "Lifecycle Management"
layout: "nsxt"
page_title: "NSXT: nsxt_policy_backup_config"
description: A resource to configure backups in NSX Policy.
---

# nsxt_policy_backup_config

This resource provides a means to configure file server and timers for automated backup.

This resource is applicable to NSX Policy Manager and Global Manager.

## Example Usage

```hcl
resource "nsxt_policy_backup_config" "backup" {
  backup_enabled = true
  weekly_backup_schedule {
    days_of_week = [0, 3, 5]
    hour_of_day = 23
    minute_of_day = 0
  }
  passphrase = "PassPhrase1:)"
  remote_file_server {
    directory_path = "/home/user"
    protocol {
      authentication_scheme {
        username = "user"
        password = "PassWord!"
        scheme_name = "PASSWORD"
      }
      ssh_fingerprint = "SHA256:xQ4gBH4llDJqg3OFSfXUL4N2iP/rTQJQI1Ew5D2kS7U"
    }
    server = "backup_server.somewhere.org"
  }
}
```

## Argument Reference

The following arguments are supported:

* `after_inventory_update_interval` - (Optional) A number of seconds after a last backup, that needs to pass, before a topology change will trigger a generation of a new cluster/node backups.
* `backup_enabled` - (Optional) true if automated backup is enabled. Default is false.
* `interval_backup_schedule` (Optional) Schedule to specify the interval time at which automated backups need to be taken.
  * `seconds_between_backups` - (Optional) Time interval in seconds between two consecutive automated backups. Default is 3600.
* `weekly_backup_schedule` - (Optional) Schedule to specify day of the week and time to take automated backup.
  * `days_of_week` - (Required) A list of days of week when backup is taken. 0 - Sunday, 1 - Monday, 2 - Tuesday, 3 - Wednesday ...
  * `hour_of_day` - (Required) Time of day when backup is taken.
  * `minute_of_day` - (Required) Time of day when backup is taken.
* `inventory_summary_interval` - (Optional) The minimum number of seconds between each upload of the inventory summary to backup server. Default is 240.
* `passphrase` - (Optional) Passphrase used to encrypt backup files.
* `remote_file_server` - (Optional) The server to which backups will be sent.
  * `directory_path` - (Required) Remote server directory to copy bundle files to.
  * `port` - (Optional) Server port. Default is 22.
  * `protocol` - (Required) Protocol to use to copy file.
    * `authentication_scheme` - (Required) Scheme to authenticate if required.
      * `identity_file` - (Optional) SSH private key data.
      * `password` - (Optional) Password to authenticate with.
      * `scheme_name` - (Required) Authentication scheme name. Valid values are: PASSWORD, KEY.
      * `username` - (Required) Username to authenticate with.
    * `protocol_name` - (Optional) Protocol name. Default is sftp.
    * `ssh_fingerprint` - (Required) SSH fingerprint of server.
  * `server` - (Required) Remote server hostname or IP address.

