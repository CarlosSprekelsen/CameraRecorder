import { WebSocketService } from '../websocket/WebSocketService';
import { LoggerService } from '../logger/LoggerService';
import { ICommand } from '../interfaces/ServiceInterfaces';

export class RecordingService implements ICommand {
  constructor(
    private wsService: WebSocketService,
    private logger: LoggerService
  ) {}

  async takeSnapshot(device: string, filename?: string): Promise<any> {
    try {
      this.logger.info('take_snapshot request', { device, filename });
      return await this.wsService.sendRPC('take_snapshot', { device, filename });
    } catch (error) {
      this.logger.error('take_snapshot failed', error as Error);
      throw error;
    }
  }

  async startRecording(device: string, duration?: number, format?: string): Promise<any> {
    try {
      this.logger.info('start_recording request', { device, duration, format });
      return await this.wsService.sendRPC('start_recording', { device, duration, format });
    } catch (error) {
      this.logger.error('start_recording failed', error as Error);
      throw error;
    }
  }

  async stopRecording(device: string): Promise<any> {
    try {
      this.logger.info('stop_recording request', { device });
      return await this.wsService.sendRPC('stop_recording', { device });
    } catch (error) {
      this.logger.error('stop_recording failed', error as Error);
      throw error;
    }
  }
}
