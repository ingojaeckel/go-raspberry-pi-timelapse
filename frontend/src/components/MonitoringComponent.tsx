import React, { useState, useEffect } from 'react';
import { apiClient, components } from '../api-client';

type MonitoringResponse = components['schemas']['MonitoringResponse'];

export default function MonitoringComponent() {
  const [state, setState] = useState<MonitoringResponse>({
    Time: "",
    Uptime: "",
    CpuTemperature: "",
    GpuTemperature: "",
    FreeDiskSpace: ""
  });

  useEffect(() => {
    apiClient
      .GET("/monitoring")
      .then(({ data }) => {
        if (data) {
          setState(data);
        }
      });
  }, []);

  return (
    <ul>
      <li>Current Time: <br/><pre>{state.Time}</pre></li>
      <li>Uptime: <br /><pre>{state.Uptime}</pre></li>
      <li>Free Disk Space: <br/><pre>{state.FreeDiskSpace}</pre></li>
      <li>Temperature:
        <ul>
          <li>GPU: <br/><pre>{state.GpuTemperature}</pre></li>
          <li>CPU: <br/><pre>{state.CpuTemperature}</pre></li>
        </ul>
      </li>
    </ul>
  );
}
