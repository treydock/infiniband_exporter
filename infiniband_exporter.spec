Name:           infiniband_exporter
Version:        0.10.0
Release:        1
Summary:        The InfiniBand exporter collects counters from InfiniBand switches and HCAs

License:        Apache License
Source0:        %{name}-%{version}.tar.gz
URL:            https://github.com/treydock/infiniband_exporter

Requires:       infiniband-diags

%description

The InfiniBand exporter collects counters from InfiniBand switches and HCAs

This exporter listens on port 9315 by default and all metrics are exposed via the /metrics endpoint.

%global debug_package %{nil}

%prep
%autosetup

%build
make build

%install
install -Dpm 0755 %{name} %{buildroot}%{_sbindir}/%{name}
install -Dpm 0644 systemd/%{name}@.service %{buildroot}%{_unitdir}/%{name}@.service
install -Dpm 0644 systemd/%{name}.sysconfig %{buildroot}%{_sysconfdir}/sysconfig/%{name}

%clean
rm -rf %{buildroot}

%pre
%{_sbindir}/useradd -c "%{name} user" -s /bin/false -r -d / %{name} 2>/dev/null || :

%files
%{_sbindir}/%{name}
%{_unitdir}/%{name}@.service
%config(noreplace) %{_sysconfdir}/sysconfig/%{name}

%changelog
* Tue Jul 2 2024 Initial RPM
- 
