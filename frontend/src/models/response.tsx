
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
	ObjectDetectionEnabled:  boolean
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

export interface BoundingBox {
    x: number
    y: number  
    width: number
    height: number
}

export interface ObjectDetail {
    class: string
    confidence: number
    category: string
    bbox?: BoundingBox
}

export interface DetectionResult {
    IsDay: boolean
    Objects: string[]
    Summary: string
    PhotoPath: string
    LatencyMs: number
    OverallConfidence: number
    Details?: ObjectDetail[]
}

export interface DetectionResponse {
    IsDay: boolean
    Objects: string[]
    Summary: string
    PhotoPath: string
    LatencyMs: number
    OverallConfidence: number
    Details?: ObjectDetail[]
}