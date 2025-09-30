/**
 * AdminPage - Architecture Compliance
 * 
 * Architecture requirement: "Admin panel for retention policies" (Priority 3)
 * Admin-only page providing system configuration interface
 * Follows existing page patterns and architecture guidelines
 */

import React from 'react';
import { Box } from '../../components/atoms/Box/Box';
import { Container } from '../../components/atoms/Container/Container';
import { Typography } from '../../components/atoms/Typography/Typography';
import { AdminPanel } from '../../components/organisms/AdminPanel/AdminPanel';
import { usePermissions } from '../../hooks/usePermissions';
import { logger } from '../../services/logger/LoggerService';

const AdminPage: React.FC = () => {
  const { canViewAdminPanel, isAdmin } = usePermissions();

  React.useEffect(() => {
    logger.info('AdminPage loaded');
  }, []);

  // Security check - redirect if not admin
  if (!canViewAdminPanel() || !isAdmin) {
    return (
      <Container maxWidth="lg">
        <Box sx={{ py: 4, textAlign: 'center' }}>
          <Typography variant="h4" color="error" gutterBottom>
            Access Denied
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Admin privileges required to access this page.
          </Typography>
        </Box>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg">
      <Box sx={{ py: 3 }}>
        <AdminPanel />
      </Box>
    </Container>
  );
};

export default AdminPage;
