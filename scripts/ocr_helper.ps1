# Standalone OCR helper — runs under Windows PowerShell 5.1 (powershell.exe).
# Loads Windows.Media.Ocr (UWP API only available via .NET Framework WinRT projection).
# Reads PNG from -ImagePath, writes recognized text to stdout.
#
# pwsh 7 cannot host WinRT projection cleanly, so the main ae_run.ps1 (pwsh 7)
# shells out to:  powershell.exe -NoProfile -File ocr_helper.ps1 -ImagePath ...
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ImagePath,
    [string]$OutFile = ''
)

$ErrorActionPreference = 'Stop'

Add-Type -AssemblyName System.Runtime.WindowsRuntime
[void][Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType=WindowsRuntime]
[void][Windows.Graphics.Imaging.SoftwareBitmap, Windows.Foundation, ContentType=WindowsRuntime]
[void][Windows.Graphics.Imaging.BitmapDecoder, Windows.Foundation, ContentType=WindowsRuntime]
[void][Windows.Storage.Streams.RandomAccessStream, Windows.Foundation, ContentType=WindowsRuntime]
[void][Windows.Storage.StorageFile, Windows.Foundation, ContentType=WindowsRuntime]
[void][Windows.Storage.FileAccessMode, Windows.Foundation, ContentType=WindowsRuntime]

$asTaskMI = [System.WindowsRuntimeSystemExtensions].GetMethods() |
    Where-Object { $_.Name -eq 'AsTask' -and $_.IsGenericMethod -and $_.GetParameters().Count -eq 1 } |
    Select-Object -First 1
if (-not $asTaskMI) { throw 'AsTask MethodInfo not found' }

function _Await($asyncOp, [type]$resultType) {
    $generic = $asTaskMI.MakeGenericMethod($resultType)
    $task    = $generic.Invoke($null, @($asyncOp))
    $task.Wait()
    return $task.Result
}

$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
if (-not $engine) {
    throw 'OCR engine unavailable -- install zh-CN / en-US language pack via Settings > Time & Language > Language'
}

$file    = _Await ([Windows.Storage.StorageFile]::GetFileFromPathAsync($ImagePath)) ([Windows.Storage.StorageFile])
$stream  = _Await $file.OpenAsync([Windows.Storage.FileAccessMode]::Read) ([Windows.Storage.Streams.IRandomAccessStream])
$decoder = _Await ([Windows.Graphics.Imaging.BitmapDecoder]::CreateAsync($stream)) ([Windows.Graphics.Imaging.BitmapDecoder])
$sb      = _Await $decoder.GetSoftwareBitmapAsync() ([Windows.Graphics.Imaging.SoftwareBitmap])
$result  = _Await ($engine.RecognizeAsync($sb)) ([Windows.Media.Ocr.OcrResult])

if ($OutFile) {
    # Write UTF-8 with no BOM to avoid mojibake when read back from pwsh.
    [System.IO.File]::WriteAllText($OutFile, $result.Text, [System.Text.UTF8Encoding]::new($false))
} else {
    Write-Output $result.Text
}
