import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { MonitoringResponse } from '../models/response'

export default function MonitoringComponent() {
  const [state, setState] = useState<MonitoringResponse>({
    Time: "",
    Uptime: "",
    CpuTemperature: "",
    GpuTemperature: "",
    FreeDiskSpace: ""
  });

  useEffect(() => {
    axios
      .get<MonitoringResponse>(BaseUrl + "/monitoring")
      .then(resp => {
        if (resp.data) {
          setState(resp.data);
        }
      });
  }, []);

  return (
    <div><h4>Monitoring</h4>
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
    </div>
  );
}
