/**
 * Network Simulation Utility for PDR-2 Testing
 * 
 * Provides network condition simulation for testing WebSocket resilience
 * Following "Real Integration First" approach with controlled network conditions
 * 
 * Used for PDR-2.1: WebSocket connection stability under network interruption
 */

export interface NetworkCondition {
  latency: number; // milliseconds
  packetLoss: number; // percentage (0-100)
  bandwidth: number; // bytes per second
  connectionStability: number; // percentage (0-100)
}

export interface NetworkScenario {
  name: string;
  description: string;
  conditions: NetworkCondition;
  duration: number; // milliseconds
}

/**
 * Predefined network scenarios for PDR-2 testing
 */
export const NETWORK_SCENARIOS: NetworkScenario[] = [
  {
    name: 'stable_connection',
    description: 'Stable network connection with minimal latency',
    conditions: {
      latency: 10,
      packetLoss: 0,
      bandwidth: 1000000, // 1MB/s
      connectionStability: 100
    },
    duration: 5000
  },
  {
    name: 'high_latency',
    description: 'High latency network (satellite connection simulation)',
    conditions: {
      latency: 500,
      packetLoss: 5,
      bandwidth: 100000, // 100KB/s
      connectionStability: 90
    },
    duration: 5000
  },
  {
    name: 'intermittent_connection',
    description: 'Intermittent connection with frequent disconnects',
    conditions: {
      latency: 100,
      packetLoss: 20,
      bandwidth: 50000, // 50KB/s
      connectionStability: 30
    },
    duration: 10000
  },
  {
    name: 'network_outage',
    description: 'Complete network outage simulation',
    conditions: {
      latency: 0,
      packetLoss: 100,
      bandwidth: 0,
      connectionStability: 0
    },
    duration: 3000
  }
];

/**
 * Network Simulator Class
 * 
 * Provides controlled network condition simulation for testing
 */
export class NetworkSimulator {
  private originalWebSocket: typeof WebSocket;
  private isSimulating = false;
  private currentScenario: NetworkScenario | null = null;

  constructor() {
    this.originalWebSocket = global.WebSocket;
  }

  /**
   * Start network simulation with specified scenario
   */
  startSimulation(scenario: NetworkScenario): void {
    if (this.isSimulating) {
      this.stopSimulation();
    }

    this.currentScenario = scenario;
    this.isSimulating = true;

    // Override WebSocket constructor to simulate network conditions
    global.WebSocket = this.createSimulatedWebSocket(scenario.conditions);

    console.log(`🌐 Network simulation started: ${scenario.name}`);
    console.log(`   Latency: ${scenario.conditions.latency}ms`);
    console.log(`   Packet Loss: ${scenario.conditions.packetLoss}%`);
    console.log(`   Bandwidth: ${scenario.conditions.bandwidth} B/s`);
    console.log(`   Stability: ${scenario.conditions.connectionStability}%`);
  }

  /**
   * Stop network simulation and restore original WebSocket
   */
  stopSimulation(): void {
    if (!this.isSimulating) {
      return;
    }

    global.WebSocket = this.originalWebSocket;
    this.isSimulating = false;
    this.currentScenario = null;

    console.log('🌐 Network simulation stopped');
  }

  /**
   * Get current simulation status
   */
  getSimulationStatus(): { isSimulating: boolean; scenario: NetworkScenario | null } {
    return {
      isSimulating: this.isSimulating,
      scenario: this.currentScenario
    };
  }

  /**
   * Create simulated WebSocket with network conditions
   */
  private createSimulatedWebSocket(conditions: NetworkCondition): typeof WebSocket {
    const simulator = this;

    return class SimulatedWebSocket extends this.originalWebSocket {
      private originalSend: (data: string | ArrayBufferLike | Blob | ArrayBufferView) => void;
      private originalClose: (code?: number, reason?: string) => void;
      private simulatedLatency: number;
      private simulatedPacketLoss: number;
      private simulatedBandwidth: number;
      private simulatedStability: number;

      constructor(url: string | URL, protocols?: string | string[]) {
        super(url, protocols);
        
        this.simulatedLatency = conditions.latency;
        this.simulatedPacketLoss = conditions.packetLoss;
        this.simulatedBandwidth = conditions.bandwidth;
        this.simulatedStability = conditions.connectionStability;

        // Override send method to simulate latency and packet loss
        this.originalSend = this.send.bind(this);
        this.send = this.simulatedSend.bind(this);

        // Override close method to simulate connection stability
        this.originalClose = this.close.bind(this);
        this.close = this.simulatedClose.bind(this);

        // Simulate connection stability issues
        if (this.simulatedStability < 100) {
          this.simulateConnectionInstability();
        }
      }

      /**
       * Simulated send with latency and packet loss
       */
      private simulatedSend(data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
        // Simulate packet loss
        if (Math.random() * 100 < this.simulatedPacketLoss) {
          console.log('📦 Simulated packet loss - message dropped');
          return;
        }

        // Simulate latency
        setTimeout(() => {
          this.originalSend(data);
        }, this.simulatedLatency);
      }

      /**
       * Simulated close with stability issues
       */
      private simulatedClose(code?: number, reason?: string): void {
        // Simulate connection stability issues
        if (Math.random() * 100 > this.simulatedStability) {
          console.log('🔌 Simulated connection instability - premature close');
          // Don't actually close, just simulate it
          return;
        }

        this.originalClose(code, reason);
      }

      /**
       * Simulate connection instability
       */
      private simulateConnectionInstability(): void {
        const instabilityInterval = setInterval(() => {
          if (Math.random() * 100 > this.simulatedStability) {
            console.log('🔌 Simulated connection drop');
            this.dispatchEvent(new Event('error'));
            this.dispatchEvent(new CloseEvent('close', { code: 1006, reason: 'Simulated network failure' }));
          }
        }, 2000 + Math.random() * 3000); // Random intervals between 2-5 seconds

        // Clean up interval when connection closes
        this.addEventListener('close', () => {
          clearInterval(instabilityInterval);
        });
      }
    };
  }

  /**
   * Run a network scenario for specified duration
   */
  async runScenario(scenario: NetworkScenario): Promise<void> {
    return new Promise((resolve) => {
      this.startSimulation(scenario);
      
      setTimeout(() => {
        this.stopSimulation();
        resolve();
      }, scenario.duration);
    });
  }

  /**
   * Run multiple scenarios sequentially
   */
  async runScenarios(scenarios: NetworkScenario[]): Promise<void> {
    for (const scenario of scenarios) {
      await this.runScenario(scenario);
      // Brief pause between scenarios
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
  }
}

/**
 * Network condition validation utility
 */
export function validateNetworkConditions(conditions: NetworkCondition): boolean {
  return (
    conditions.latency >= 0 &&
    conditions.latency <= 10000 &&
    conditions.packetLoss >= 0 &&
    conditions.packetLoss <= 100 &&
    conditions.bandwidth >= 0 &&
    conditions.bandwidth <= 10000000 &&
    conditions.connectionStability >= 0 &&
    conditions.connectionStability <= 100
  );
}

/**
 * Create custom network scenario
 */
export function createCustomScenario(
  name: string,
  description: string,
  conditions: NetworkCondition,
  duration: number
): NetworkScenario {
  if (!validateNetworkConditions(conditions)) {
    throw new Error('Invalid network conditions provided');
  }

  return {
    name,
    description,
    conditions,
    duration
  };
}
