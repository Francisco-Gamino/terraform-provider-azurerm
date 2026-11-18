change "breaking" {
  body = "`azurerm_function_app_flex_consumption` - reject PowerShell `runtime_version` values `7`, `7.0`, `7.2`, and `7.4`; use `7.6` or another version supported by Azure Functions"
}

change "breaking" {
  body = "`azurerm_linux_function_app` - remove support for PowerShell `7`, `7.2`, and `7.4` in `site_config.application_stack.powershell_core_version`; use `7.6`"
}

change "breaking" {
  body = "`azurerm_linux_function_app_slot` - remove support for PowerShell `7`, `7.2`, and `7.4` in `site_config.application_stack.powershell_core_version`; use `7.6`"
}

change "breaking" {
  body = "`azurerm_windows_function_app` - remove support for PowerShell `7`, `7.2`, and `7.4` in `site_config.application_stack.powershell_core_version`; use `7.6`"
}

change "breaking" {
  body = "`azurerm_windows_function_app_slot` - remove support for PowerShell `7`, `7.2`, and `7.4` in `site_config.application_stack.powershell_core_version`; use `7.6`"
}
