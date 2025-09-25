import React from 'react';
import { Box, Tabs, Tab, Badge } from '@mui/material';
import { VideoLibrary as RecordingsIcon, PhotoLibrary as SnapshotsIcon } from '@mui/icons-material';

interface FileTabsProps {
  currentTab: 'recordings' | 'snapshots';
  onTabChange: (tab: 'recordings' | 'snapshots') => void;
  recordingsCount: number;
  snapshotsCount: number;
}

/**
 * FileTabs - Tab navigation for recordings and snapshots
 * Following architecture section 5.1 for file management UI
 */
const FileTabs: React.FC<FileTabsProps> = ({
  currentTab,
  onTabChange,
  recordingsCount,
  snapshotsCount,
}) => {
  const handleChange = (_event: React.SyntheticEvent, newValue: 'recordings' | 'snapshots') => {
    onTabChange(newValue);
  };

  return (
    <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
      <Tabs value={currentTab} onChange={handleChange} aria-label="file tabs">
        <Tab
          icon={
            <Badge badgeContent={recordingsCount} color="primary">
              <RecordingsIcon />
            </Badge>
          }
          iconPosition="start"
          label="Recordings"
          value="recordings"
        />
        <Tab
          icon={
            <Badge badgeContent={snapshotsCount} color="primary">
              <SnapshotsIcon />
            </Badge>
          }
          iconPosition="start"
          label="Snapshots"
          value="snapshots"
        />
      </Tabs>
    </Box>
  );
};

export default FileTabs;
