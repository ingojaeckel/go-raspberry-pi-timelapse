import React from 'react';
import axios from 'axios';
import { Box, Container, CssBaseline, Grid, Tab, Tabs, Typography, useMediaQuery, useTheme } from '@mui/material';
import { ThemeProvider } from '@mui/material/styles';
import { Home, PhotoCamera, Timeline, Settings, Description } from '@mui/icons-material';
import PhotoComponent from './components/PhotoComponent';
import MonitoringComponent from './components/MonitoringComponent';
import SetupComponent from './components/SetupComponent';
import PreviewComponent from './components/PreviewComponent';
import LogComponent from './components/LogComponent';
import theme from './theme';

// Register default request interceptors
axios.interceptors.request.use(request => {
  console.log('Starting Request', JSON.stringify(request, null, 2))
  return request
})

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ p: 3 }}>
          <Typography component={'span'}>{children}</Typography>
        </Box>
      )}
    </div>
  );
}

function App() {
  const [value, setValue] = React.useState(0);
  const muiTheme = useTheme();
  const isMobile = useMediaQuery(muiTheme.breakpoints.down('sm'));

  const handleChange = (_event: React.SyntheticEvent, newValue: number) => {
    setValue(newValue);
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Container maxWidth="xl">
        <Box sx={{ width: '100%', mt: 2 }}>
          <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
            <Tabs
              value={value}
              onChange={handleChange}
              variant={isMobile ? "scrollable" : "standard"}
              scrollButtons={isMobile ? "auto" : false}
              aria-label="basic tabs"
              sx={{
                '& .MuiTab-root': {
                  minWidth: isMobile ? 'auto' : 120,
                  py: isMobile ? 1 : 2,
                }
              }}
            >
              <Tab icon={<Home />} label={!isMobile && "Home"} />
              <Tab icon={<PhotoCamera />} label={!isMobile && "Photo"} />
              <Tab icon={<Timeline />} label={!isMobile && "Monitoring"} />
              <Tab icon={<Settings />} label={!isMobile && "Setup"} />
              <Tab icon={<Description />} label={!isMobile && "Logs"} />
            </Tabs>
          </Box>
          <Box sx={{ mt: 2 }}>
            <TabPanel value={value} index={0}>
              <PreviewComponent />
            </TabPanel>
            <TabPanel value={value} index={1}>
              <PhotoComponent />
            </TabPanel>
            <TabPanel value={value} index={2}>
              <MonitoringComponent />
            </TabPanel>
            <TabPanel value={value} index={3}>
              <SetupComponent />
            </TabPanel>
            <TabPanel value={value} index={4}>
              <LogComponent />
            </TabPanel>
          </Box>
        </Box>
      </Container>
    </ThemeProvider>
  );
}

export default App;
