import React from 'react';
import { Box } from '../atoms/Box/Box';
import { Tabs, Tab } from '../atoms/Tabs/Tabs';
import { Badge } from '../atoms/Badge/Badge';
import { Icon } from '../atoms/Icon/Icon';

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
  const handleChange = (_event: React.SyntheticEvent, newValue: number) => {
    const tabValue = newValue === 0 ? 'recordings' : 'snapshots';
    onTabChange(tabValue);
  };

  const currentIndex = currentTab === 'recordings' ? 0 : 1;

  return (
    <Box className="border-b border-gray-200">
      <Tabs value={currentIndex} onChange={handleChange}>
        <Tab
          label={
            <div className="flex items-center gap-2">
              <Badge badgeContent={recordingsCount} color="primary">
                <Icon name="recordings" size={16} />
              </Badge>
              <span>Recordings</span>
            </div>
          }
        />
        <Tab
          label={
            <div className="flex items-center gap-2">
              <Badge badgeContent={snapshotsCount} color="primary">
                <Icon name="camera" size={16} />
              </Badge>
              <span>Snapshots</span>
            </div>
          }
        />
      </Tabs>
    </Box>
  );
};

export default FileTabs;
