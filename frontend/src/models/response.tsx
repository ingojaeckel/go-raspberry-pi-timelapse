
export interface MonitoringResponse {
    Time: string
    Uptime: string
    CpuTemperature: string
    GpuTemperature: string
    FreeDiskSpace: string
}

export interface SettingsResponse {
	SecondsBetweenCaptures:  number
	OffsetWithinHour:        number
	PhotoResolutionWidth:    number
	PhotoResolutionHeight:   number
	PreviewResolutionWidth:  number
	PreviewResolutionHeight: number
	RotateBy:                number
	ResolutionSetting:       number
	Quality:                 number
	DebugEnabled:            boolean
}

export interface LogResponse {
	Logs: string
}

export interface PhotosResponse {
    Photos: Photo[]
}

export interface Photo {
    Name: string
    ModTime: string
    Size: string
}