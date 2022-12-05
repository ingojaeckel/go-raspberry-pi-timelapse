import React from 'react';
import { createBrowserHistory } from "history";
import { NavLink, Router } from 'react-router-dom';
import Switcher from './Switch';
import axios from 'axios';
import { Avatar, Container, CssBaseline, Grid, List, ListItem, ListItemAvatar, Tab, Tabs } from '@mui/material';
import { Home, PhotoCamera, Timeline, Settings, Description } from '@mui/icons-material'

// Register default request interceptors
axios.interceptors.request.use(request => {
  console.log('Starting Request', JSON.stringify(request, null, 2))
  return request
})

const customHistory = createBrowserHistory();

interface LinkTabProps {
  label: string;
  href: string;
  icon: React.ReactElement;
}

function LinkTab(props: LinkTabProps) {
  return (
    <Tab
      component="a"
      onClick={(event: React.MouseEvent<HTMLAnchorElement, MouseEvent>) => {
        //event.preventDefault();
      }}
      {...props}
    />
  );
}

function App() {
  const [value, setValue] = React.useState(0);
  const handleChange = (event: React.SyntheticEvent, newValue: number) => {
    setValue(newValue);
  };
  return (
    <Router history={customHistory}>
      <React.Fragment>
        <CssBaseline />
        <Container fixed>
          <Grid container spacing={0}>
            <Grid item xs={12}>
              <Tabs value={value} onChange={handleChange} aria-label="icon label tabs">
                <LinkTab label="Home" href={`${process.env.PUBLIC_URL}/`} icon={<Home />} />
                <LinkTab label="Preview" href={`${process.env.PUBLIC_URL}/preview`} icon={<PhotoCamera />} />
                <LinkTab label="Monitoring" href={`${process.env.PUBLIC_URL}/monitoring`} icon={<Timeline />} />
                <LinkTab label="Settings" href={`${process.env.PUBLIC_URL}/settings`} icon={<Settings />} />
                <LinkTab label="Logs" href={`${process.env.PUBLIC_URL}/logs`} icon={<Description />} />
              </Tabs>
            </Grid>
            <Grid item xs={10}>
              <Switcher />
            </Grid>
            <Grid item xs={12}>
              <div className="footer">version: <a href={"https://github.com/ingojaeckel/go-raspberry-pi-timelapse/commit/" + process.env.REACT_APP_GIT_SHA}>{process.env.REACT_APP_GIT_SHA}</a></div>
            </Grid>
          </Grid>
        </Container>
      </React.Fragment>
    </Router>
  );
}

export default App;
