package powershellexec

// ===== LICENSE FOR WRAPPER SCRIPT FROM github.com/chef/chef =====
//
// Author:: Adam Edwards (<adamed@chef.io>)
// Copyright:: Copyright 2013-2016, Chef Software Inc.
// License:: Apache License, Version 2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// PowershellWrapperScript is a wrapper script that handles some setup
// and wrapping behaviors that are useful for most usecases.
var PowershellWrapperScript = `
# Chef Client wrapper for powershell_script resources
# In rare cases, such as when PowerShell is executed
# as an alternate user, the new-variable cmdlet is not
# available, so import it just in case
if ( get-module -ListAvailable Microsoft.PowerShell.Utility )
{
	Import-Module Microsoft.PowerShell.Utility
}
# LASTEXITCODE can be uninitialized -- make it explictly 0
# to avoid incorrect detection of failure (non-zero) codes
$global:LASTEXITCODE = 0
# Catch any exceptions -- without this, exceptions will result
# In a zero return code instead of the desired non-zero code
# that indicates a failure
trap [Exception] {write-error ($_.Exception.Message);exit 1}
# Variable state that should not be accessible to the user code
new-variable -name interpolatedexitcode -visibility private -value $true
new-variable -name chefscriptresult -visibility private
# Initialize a variable we use to capture $? inside a block
$global:lastcmdlet = $null
# Execute the user's code in a script block --
$chefscriptresult =
{

SCRIPTBLOCK

 # This assignment doesn't affect the block's return value
 $global:lastcmdlet = $?
}.invokereturnasis()
# Assume failure status of 1 -- success cases
# will have to override this
$exitstatus = 1
# If convert_boolean_return is enabled, the block's return value
# gets precedence in determining our exit status
if ($interpolatedexitcode -and $chefscriptresult -ne $null -and $chefscriptresult.gettype().name -eq 'boolean')
{
  $exitstatus = [int32](!$chefscriptresult)
}
elseif ($lastcmdlet)
{
  # Otherwise, a successful cmdlet execution defines the status
  $exitstatus = 0
}
elseif ( $LASTEXITCODE -ne $null -and $LASTEXITCODE -ne 0 )
{
  # If the cmdlet status is failed, allow the Win32 status
  # in $LASTEXITCODE to define exit status. This handles the case
  # where no cmdlets, only Win32 processes have run since $?
  # will be set to $false whenever a Win32 process returns a non-zero
  # status.
  $exitstatus = $LASTEXITCODE
}
# Print STDOUT for the script execution
Write-Output $chefscriptresult
# If this script is launched with -File, the process exit
# status of PowerShell.exe will be $exitstatus. If it was
# launched with -Command, it will be 0 if $exitstatus was 0,
# 1 (i.e. failed) otherwise.
exit $exitstatus
`
