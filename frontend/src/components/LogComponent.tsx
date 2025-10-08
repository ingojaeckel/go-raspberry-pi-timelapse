import React, { useState, useEffect } from 'react';
import { Button, ButtonGroup } from '@mui/material';
import { apiClient } from '../api-client';

export default function LogComponent() {
  const [logs, setLogs] = useState<string>("");

  const getLogs = () => {
    console.log("Getting logs...");
    apiClient
      .GET("/logs")
      .then(({ data }) => {
        if (data) {
          setLogs(data.Logs);
        }
      });
  }

  useEffect(() => getLogs(), []);

  return (
    <React.Fragment>
      <ButtonGroup color="primary" aria-label="outlined primary button group">
        <Button onClick={() => getLogs()}>Refresh</Button>
      </ButtonGroup>
      <div><pre>{logs}</pre></div>
    </React.Fragment>
  );
}
