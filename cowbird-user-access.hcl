# cowbird-user-access — per-user Vault policy for cowbird clients.
#
# Uses {{identity.entity.id}} templating; assign directly to users
# (userpass in Vault 2.0.0 emits no group claims, so group-based
# auto-enrollment does not work).
#
# Gotcha verified empirically: Vault does NOT merge capabilities across
# matching rules — the most specific matching path wins outright. Any exact
# templated path shadowed by a glob must repeat every capability the glob
# would have granted (see the own-pubkey rule below).

# --- own subtree: items, identity, links, share records -----------------------

path "cowbird/data/users/{{identity.entity.id}}/*" {
  capabilities = ["create", "read", "update", "delete"]
}

path "cowbird/metadata/users/{{identity.entity.id}}/*" {
  capabilities = ["list", "delete"]
}

# --- public-key directory: read/list all, write own ---------------------------

path "cowbird/data/pubkeys/*" {
  capabilities = ["read"]
}

# Listing the directory hits the metadata path of the prefix itself.
path "cowbird/metadata/pubkeys" {
  capabilities = ["list"]
}

# Exact match shadows the data/pubkeys/* glob, so "read" must be repeated
# here or reading one's own entry is denied.
path "cowbird/data/pubkeys/{{identity.entity.id}}" {
  capabilities = ["create", "read", "update"]
}

# --- shared envelopes: read all, owner-managed --------------------------------
# Path ACL is not the security boundary here; the wrapped item key is.
# Broad read is safe because contents are encrypted.

path "cowbird/data/shared/*" {
  capabilities = ["read"]
}

path "cowbird/data/shared/{{identity.entity.id}}/*" {
  capabilities = ["create", "read", "update", "delete"]
}

path "cowbird/metadata/shared/{{identity.entity.id}}/*" {
  capabilities = ["list", "delete"]
}

# --- inbox: recipient reads/deletes own; senders may only create --------------
# create-without-update lets a sender drop a new message but not overwrite,
# read, or list. The templated own-inbox rules take precedence over the
# sender wildcard for one's own inbox.

path "cowbird/data/inbox/{{identity.entity.id}}/*" {
  capabilities = ["read", "delete"]
}

path "cowbird/metadata/inbox/{{identity.entity.id}}/*" {
  capabilities = ["list", "delete"]
}

path "cowbird/data/inbox/+/*" {
  capabilities = ["create"]
}
