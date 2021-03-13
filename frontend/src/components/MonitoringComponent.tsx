import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { MonitoringResponse } from '../models/response'
import { Typography } from '@material-ui/core';

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
    <div>
      <div>
        <Typography variant="h4" component="h4">Monitoring</Typography>
      </div>
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
