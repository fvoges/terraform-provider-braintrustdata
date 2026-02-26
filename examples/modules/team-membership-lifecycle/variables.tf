variable "identity_case_insensitive" {
  description = "Normalize identity keys (for example emails) to lowercase before resolution."
  type        = bool
  default     = true
}

variable "team_group_name_prefix" {
  description = "Optional prefix added to every created team group name."
  type        = string
  default     = ""
}

variable "binding_key_delimiter" {
  description = "Delimiter used in output binding keys. Must match project-access-lifecycle delimiter."
  type        = string
  default     = "|"

  validation {
    condition     = trimspace(var.binding_key_delimiter) != ""
    error_message = "binding_key_delimiter must not be empty."
  }

  validation {
    condition     = var.binding_key_delimiter == "|"
    error_message = "binding_key_delimiter must be \"|\" to remain compatible with project-access-lifecycle."
  }
}

variable "resolve_identities_with_data_source" {
  description = "When true, resolve missing identities via data.braintrustdata_user lookup by email."
  type        = bool
  default     = true
}

variable "org_name_for_user_lookup" {
  description = "Optional org_name passed to braintrustdata_user lookup."
  type        = string
  default     = null
}

variable "user_id_by_identity" {
  description = "Optional map of identity key -> user ID for deterministic/offline resolution."
  type        = map(string)
  default     = {}
}

variable "teams" {
  description = "Team catalog keyed by stable team key."
  type = map(object({
    name              = optional(string)
    description       = optional(string)
    member_identities = optional(set(string), [])
    member_user_ids   = optional(set(string), [])
    member_team_keys  = optional(set(string), [])
    project_role_bindings = optional(set(object({
      project_key = string
      role_key    = string
    })), [])
  }))
  default = {}
}
