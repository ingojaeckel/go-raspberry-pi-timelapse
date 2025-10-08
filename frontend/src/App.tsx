import React from 'react';
import axios from 'axios';
import { Box, Container, CssBaseline, Grid, Tab, Tabs, Typography } from '@mui/material';
import { Home, PhotoCamera, Timeline, Settings, Description } from '@mui/icons-material'
import PhotoComponent from './components/PhotoComponent';
import MonitoringComponent from './components/MonitoringComponent';
import SetupComponent from './components/SetupComponent';
import PreviewComponent from './components/PreviewComponent';
import LogComponent from './components/LogComponent';

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
        <Box sx={{ p: { xs: 1, sm: 3 } }}>
          <Typography component={'span'}>{children}</Typography>
        </Box>
      )}
    </div>
  );
}

function App() {
  const [value, setValue] = React.useState(0);
  const handleChange = (_event: React.SyntheticEvent, newValue: number) => {
    setValue(newValue);
  };
  return (
    <React.Fragment>
      <CssBaseline />
      <Container 
        sx={{ 
          padding: { xs: '0', sm: '16px' },
          maxWidth: { xs: '100%', sm: 'lg' }
        }}
      >
        <Grid container spacing={0}>
          <Grid size={12}>
            <Tabs 
              value={value} 
              onChange={handleChange} 
              variant="scrollable" 
              scrollButtons 
              allowScrollButtonsMobile 
              aria-label="scrollable icon label tabs"
              sx={{
                minHeight: { xs: '48px', sm: '64px' },
                '& .MuiTab-root': {
                  minHeight: { xs: '48px', sm: '64px' },
                  padding: { xs: '6px 8px', sm: '12px 16px' },
                  fontSize: { xs: '0.75rem', sm: '0.875rem' },
                  minWidth: { xs: '60px', sm: '90px' },
                  '& .MuiSvgIcon-root': {
                    fontSize: { xs: '1.2rem', sm: '1.5rem' },
                  },
                },
              }}
            >
              <Tab label="home" icon={<Home />} />
              <Tab label="preview" icon={<PhotoCamera />} />
              <Tab label="monitoring" icon={<Timeline />} />
              <Tab label="settings" icon={<Settings />} />
              <Tab label="logs" icon={<Description />} />
            </Tabs>
            <TabPanel value={value} index={0}><PhotoComponent /></TabPanel>
            <TabPanel value={value} index={1}><PreviewComponent /></TabPanel>
            <TabPanel value={value} index={2}><MonitoringComponent /></TabPanel>
            <TabPanel value={value} index={3}><SetupComponent /></TabPanel>
            <TabPanel value={value} index={4}><LogComponent /></TabPanel>
          </Grid>
        </Grid>
      </Container>
    </React.Fragment>
  );
}

export default App;
