data "acloud_cloud_profile" "cyso" {
  name = "Cyso Cloud AMS2"
}

resource "acloud_cloud_account" "cyso" {
  display_name           = "Cyso Cloud AMS2"
  cloud_profile_identity = data.acloud_cloud_profile.cyso.identity
  openstack_tenant_id    = "0123456789abcdef0123456789abcdef"
}
