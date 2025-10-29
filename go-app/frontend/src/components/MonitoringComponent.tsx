import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config'
import { MonitoringResponse } from '../models/response'
import { Alert, Card, CardContent, Grid, Typography, Box, CircularProgress } from '@mui/material';
import { Schedule, Storage, Thermostat } from '@mui/icons-material';

export default function MonitoringComponent() {
  const [state, setState] = useState<MonitoringResponse>({
    Time: "",
    Uptime: "",
    CpuTemperature: "",
    GpuTemperature: "",
    FreeDiskSpace: ""
  });
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>("");

  useEffect(() => {
    setLoading(true);
    setError("");
    axios
      .get<MonitoringResponse>(BaseUrl + "/monitoring")
      .then(resp => {
        if (resp.data) {
          setState(resp.data);
          setLoading(false);
        }
      })
      .catch(err => {
        setLoading(false);
        if (err.code === 'ERR_NETWORK' || err.message.includes('Network Error')) {
          setError("Unable to connect to server. Please ensure the server is running.");
        } else {
          setError("Failed to fetch monitoring data: " + (err.message || "Unknown error"));
        }
      });
  }, []);

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight={200}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error" sx={{ mb: 2 }}>
        {error}
      </Alert>
    );
  }

  return (
    <Grid container spacing={2}>
      <Grid size={12}>
        <Card>
          <CardContent>
            <Box display="flex" alignItems="center" mb={1}>
              <Schedule sx={{ mr: 1 }} />
              <Typography variant="h6" component="div">
                System Time
              </Typography>
            </Box>
            <Typography variant="body1" color="text.secondary">
              {state.Time || "N/A"}
            </Typography>
          </CardContent>
        </Card>
      </Grid>
      
      <Grid size={12}>
        <Card>
          <CardContent>
            <Box display="flex" alignItems="center" mb={1}>
              <Schedule sx={{ mr: 1 }} />
              <Typography variant="h6" component="div">
                Uptime
              </Typography>
            </Box>
            <Typography variant="body1" color="text.secondary">
              {state.Uptime || "N/A"}
            </Typography>
          </CardContent>
        </Card>
      </Grid>

      <Grid size={12}>
        <Card>
          <CardContent>
            <Box display="flex" alignItems="center" mb={1}>
              <Storage sx={{ mr: 1 }} />
              <Typography variant="h6" component="div">
                Free Disk Space
              </Typography>
            </Box>
            <Typography variant="body1" color="text.secondary">
              {state.FreeDiskSpace || "N/A"}
            </Typography>
          </CardContent>
        </Card>
      </Grid>

      <Grid size={6}>
        <Card>
          <CardContent>
            <Box display="flex" alignItems="center" mb={1}>
              <Thermostat sx={{ mr: 1 }} />
              <Typography variant="h6" component="div">
                CPU Temperature
              </Typography>
            </Box>
            <Typography variant="body1" color="text.secondary">
              {state.CpuTemperature || "N/A"}
            </Typography>
          </CardContent>
        </Card>
      </Grid>

      <Grid size={6}>
        <Card>
          <CardContent>
            <Box display="flex" alignItems="center" mb={1}>
              <Thermostat sx={{ mr: 1 }} />
              <Typography variant="h6" component="div">
                GPU Temperature
              </Typography>
            </Box>
            <Typography variant="body1" color="text.secondary">
              {state.GpuTemperature || "N/A"}
            </Typography>
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  );
}
