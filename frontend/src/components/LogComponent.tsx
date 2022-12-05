import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Button, ButtonGroup, Typography } from '@mui/material';
import { LogResponse } from '../models/response'
import { BaseUrl } from '../conf/config'

export default function LogComponent() {
  const [logs, setLogs] = useState<string>("");

  const getLogs = () => {
    console.log("Getting logs...");
    axios
      .get<LogResponse>(BaseUrl + "/logs")
      .then(resp => {
        if (resp.data) {
          setLogs(resp.data.Logs);
        }
      });
  }

  useEffect(() => getLogs(), []);

  return (
    <React.Fragment>
      <Typography variant="h4" component="h4">Logs</Typography>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={() => getLogs()}>Refresh</Button>
      </ButtonGroup>
      <div><pre>{logs}</pre></div>
    </React.Fragment>
  );
}
